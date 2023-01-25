package mosaic

import (
	"container/heap"
	"crypto/rand"
	"image"
	"image/draw"
	"log"
	"math"
	"math/big"
	"os"

	"github.com/Woolfno/mosaic/color"
	"github.com/Woolfno/mosaic/queue"
	"github.com/Woolfno/mosaic/store"
	"github.com/Woolfno/mosaic/workerpool"

	"github.com/nfnt/resize"
)

const PAZZLE_SIZE = 32

type Mosaic struct {
	img       image.Image
	frameSize int
	store     []store.Store
}

type taskArgs struct {
	x  int
	y  int
	kx int
	ky int
}

type taskResult struct {
	kx   int
	ky   int
	path string
}

func New(img image.Image, frameSize int, data []store.Store) *Mosaic {
	return &Mosaic{
		img:       img,
		frameSize: frameSize,
		store:     data,
	}
}

func (p *Mosaic) GeneratePuzzles() (image.Image, error) {
	bounds := p.img.Bounds()
	newImg := image.NewRGBA(
		image.Rect(
			0,
			0,
			bounds.Dx()/p.frameSize*PAZZLE_SIZE,
			bounds.Dy()/p.frameSize*PAZZLE_SIZE,
		),
	)

	kx := 0
	ky := 0
	offset := PAZZLE_SIZE - p.frameSize
	for y := bounds.Min.Y; y < bounds.Max.Y; y += p.frameSize {
		kx = 0
		for x := bounds.Min.X; x < bounds.Max.X; x += p.frameSize {
			avg := AverageColorFrame(p.img, x, y, p.frameSize)
			img := SimilarImage(avg, p.store)
			if err := Draw(newImg, img, x+kx, y+ky, p.frameSize); err != nil {
				return nil, err
			}
			kx += offset
		}
		ky += offset
	}

	return newImg, nil
}

func (p *Mosaic) GeneratePuzzlesMultithread() (image.Image, error) {
	var tasks []*workerpool.Task

	bounds := p.img.Bounds()
	newImg := image.NewRGBA(
		image.Rect(
			0,
			0,
			bounds.Dx()/p.frameSize*PAZZLE_SIZE,
			bounds.Dy()/p.frameSize*PAZZLE_SIZE,
		),
	)

	var kx int
	var ky int
	offset := PAZZLE_SIZE - p.frameSize
	for y := bounds.Min.Y; y < bounds.Max.Y; y += p.frameSize {
		kx = 0
		for x := bounds.Min.X; x < bounds.Max.X; x += p.frameSize {
			task := workerpool.NewTask(func(args interface{}) (interface{}, error) {
				tArgs := args.(taskArgs)
				avg := AverageColorFrame(p.img, tArgs.x, tArgs.y, p.frameSize)
				path := SimilarImage(avg, p.store)
				return taskResult{kx: tArgs.x + tArgs.kx, ky: tArgs.y + tArgs.ky, path: path}, nil
			}, taskArgs{
				x: x, y: y, kx: kx, ky: ky,
			})
			tasks = append(tasks, task)

			kx += offset
		}
		ky += offset
	}

	pool := workerpool.NewPool(tasks, 10)
	pool.Run()

	for _, t := range tasks {
		result := t.Result.(taskResult)
		if err := Draw(newImg, result.path, result.kx, result.ky, p.frameSize); err != nil {
			return nil, err
		}
	}
	return newImg, nil
}

// средний цвет на участке frameSize * frameSize
func AverageColorFrame(img image.Image, x int, y int, frameSize int) color.Color {
	return AverageColor(img, image.Rect(x, y, x+frameSize, y+frameSize))
}

// среднеий цвет картинки
func AverageColorImage(img image.Image) color.Color {
	return AverageColor(img, img.Bounds())
}

func AverageColor(img image.Image, r image.Rectangle) color.Color {
	var rgb color.Color
	count := r.Dx() * r.Dy()

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgb.R += r >> 8 // >>8 для RGBA в RGB
			rgb.G += g >> 8
			rgb.B += b >> 8
		}
	}
	rgb.R = rgb.R / uint32(count)
	rgb.G = rgb.G / uint32(count)
	rgb.B = rgb.B / uint32(count)
	return rgb
}

// манхэттенское расстояние для поиска похожей картинки
func Distance(a color.Color, b color.Color) uint32 {
	d := 0.0
	d += math.Abs(float64(a.R) - float64(b.R))
	d += math.Abs(float64(a.G) - float64(b.G))
	d += math.Abs(float64(a.B) - float64(b.B))
	return uint32(d)
}

// определение наиболее подходящей картинки
// по минимальному расстоянию между каждым каналом цвета пикселя исходной картинки
// и средним цветом искомой картинки
// random_mode - случайная из 5 подходящих
func SimilarImage(point color.Color, store []store.Store) string {
	distanceStore := make(queue.PriorityQueue, 0, len(store))
	randomMod := true
	randomCount := 5

	for _, rgb := range store {
		d := Distance(point, *color.New(uint32(rgb.R), uint32(rgb.G), uint32(rgb.B)))
		distanceStore.Push(queue.Item{Key: rgb.Path, Value: d})
	}

	heap.Init(&distanceStore)

	n := 0
	if randomMod {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(randomCount)))
		if err != nil {
			log.Fatal(err)
		}
		n = int(r.Int64())
	}

	item := heap.Pop(&distanceStore)
	for ; n > 0; n-- {
		item = heap.Pop(&distanceStore)
	}
	return item.(*queue.Item).Key
}

// вставка картинки
func Draw(dst draw.Image, src string, x int, y int, frame_size int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	draw.Draw(dst,
		image.Rect(x, y, x+PAZZLE_SIZE, y+PAZZLE_SIZE),
		img,
		image.Point{0, 0},
		draw.Src)
	return nil
}

// Изменение размера
func Resize(original image.Image, size uint) image.Image {
	return resize.Resize(size, size, original, resize.Lanczos2)
}
