package wikimedia

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mosaic/wikimedia/response"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const API_URL = "https://en.wikipedia.org/api/rest_v1/"

type Wikimedia struct {
	url string
}

func New() *Wikimedia {
	return &Wikimedia{url: API_URL}
}

// Получить заголовок случайной страницы
func (w *Wikimedia) RandomPageTitle() string {
	url := w.url + "page/random/title"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var items response.RandomPageTitle

	if err := json.Unmarshal(body, &items); err != nil {
		log.Fatal(err)
	}

	return items.Items[0]["title"].(string)
}

// Получить список картинок, используемых на странице.
func (w *Wikimedia) MediaList(title string) []string {
	url := fmt.Sprintf("%spage/media-list/%s", w.url, title)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var items response.MediaList
	if err := json.Unmarshal(body, &items); err != nil {
		log.Fatal(err)
	}

	var images []string
	for _, item := range items.Items {
		if item.Type == "image" {
			if len(item.SrcSet) == 0 {
				continue
			}
			imgSrc := item.SrcSet[0].Src
			for _, src := range item.SrcSet {
				if src.Scale == "1x" {
					imgSrc = src.Src
				}
			}
			images = append(images, "https:"+imgSrc)
		}
	}

	return images
}

// Загрузка изображение
func (w *Wikimedia) Upload(url string) ([]byte, error) {
	var resp *http.Response
	var err error
	for repeats := 5; repeats > 0; repeats-- {
		resp, err = http.Get(url)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			repeats = 0
		case http.StatusTooManyRequests:
			log.Printf("%s: %s. Wait 1 Seconds", resp.Status, url)
			time.Sleep(1 * time.Second)
		default:
			log.Printf("%s: %s.", resp.Status, url)
			return nil, fmt.Errorf("upload error: %v", resp.StatusCode)
		}
	}

	if resp.StatusCode == http.StatusOK {
		return io.ReadAll(resp.Body)
	}
	return nil, fmt.Errorf("upload error: %v", resp.Status)
}

func worker(wg *sync.WaitGroup, urls <-chan string, dstDir string, processing func(path string) error) {
	defer wg.Done()
	w := Wikimedia{}

	for url := range urls {
		img, err := w.Upload(url)
		if err != nil {
			log.Println(err)
			continue
		}

		path := filepath.Join(dstDir, filepath.Base(url))
		if err := w.Save(path, img); err != nil {
			log.Println(err)
			continue
		}

		if err := processing(path); err != nil {
			log.Println(err)
		}
	}
}

// Полечение случайной картинки с использованием goroutines
func (w *Wikimedia) AsyncLoadRandomImage(dstDir string, count int, processing func(path string) error) error {
	urls := make(chan string, 100)
	poolSize := 4
	var wg sync.WaitGroup

	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go worker(&wg, urls, dstDir, processing)
	}

	for count > 0 {
		title := w.RandomPageTitle()
		paths := w.MediaList(title)
		for i := 0; i < len(paths) && count > 0; i++ {
			urls <- paths[i]
			count--
		}
	}
	close(urls)
	wg.Wait()

	return nil
}

// Полечение случайной картинки
func (w *Wikimedia) LoadRandomImage(dstDir string, count int, processing func(path string) error) error {
	log.Printf("Load %d random images from wikimedia.", count)

	for count > 0 {
		title := w.RandomPageTitle()
		paths := w.MediaList(title)
		for i := 0; i < len(paths) && count > 0; i++ {
			url := paths[i]

			img, err := w.Upload(url)
			if err != nil {
				log.Fatal(err)
			}

			path := filepath.Join(dstDir, filepath.Base(url))
			if err := w.Save(path, img); err != nil {
				log.Fatal(err)
			}

			if err := processing(path); err != nil {
				log.Fatal(err)
			}

			count--
		}

	}

	return nil
}

func (w *Wikimedia) Save(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
