package processing

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"mosaic/mosaic"
	"mosaic/store"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func worker(wg *sync.WaitGroup, paths <-chan string, outDir string, results chan<- error) {
	defer wg.Done()
	for path := range paths {
		img, err := LoadImage(path)
		if err != nil {
			results <- err
			continue
		}
		newImg := mosaic.Resize(img, 32)
		if err := SaveImage(newImg, filepath.Join(outDir, filepath.Base(path))); err != nil {
			results <- err
		}
	}
}

func MakeSmallImages(inDir string, outDir string) error {
	paths := make(chan string, 100)
	results := make(chan error, 100)
	poolSize := 2
	var wg sync.WaitGroup

	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go worker(&wg, paths, outDir, results)
	}

	err := filepath.Walk(inDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".jpeg" || ext == ".jpg" {
				paths <- path
			}
		}
		return nil
	})
	close(paths)
	wg.Wait()
	close(results)

	for result := range results {
		if result != nil {
			log.Println(result)
		}
	}

	return err
}

func ResizeImage(path string, outDir string, size uint) error {
	img, err := LoadImage(path)
	if err != nil {
		return err
	}

	newImg := mosaic.Resize(img, size)
	return SaveImage(newImg, filepath.Join(outDir, filepath.Base(path)))
}

func LoadImage(filename string) (image.Image, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open:%s, %v", filename, err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode '%s', %v", filename, err)
	}
	return img, nil
}

func SaveImage(img image.Image, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	jpeg.Encode(out, img, &jpeg.Options{Quality: 80})
	return nil
}

func CalculateAverage(imgDir string) (*[]store.Store, error) {
	log.Printf("Calculate average color images from '%s'.", imgDir)

	var data []store.Store
	exts := []string{"jpeg", "jpg", "png"}
	err := filepath.Walk(imgDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		isImg := false
		for _, e := range exts {
			if ext == e {
				isImg = true
				break
			}
		}
		if isImg {
			img, err := LoadImage(path)
			if err != nil {
				return err
			}

			avg := mosaic.AverageColorImage(img)
			data = append(data, store.Store{
				Path: path,
				R:    uint(avg.R),
				G:    uint(avg.G),
				B:    uint(avg.B)},
			)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &data, nil
}
