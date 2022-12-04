package matrix

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
)

type M struct {
	pixels [][]float64
	bounds image.Rectangle
}

func New(bounds image.Rectangle) *M {
	dx, dy := bounds.Dx(), bounds.Dy()
	picture := make([][]float64, dy)
	pixels := make([]float64, dx*dy)
	for i := range picture {
		picture[i], pixels = pixels[:dx], pixels[dx:]
	}

	m := new(M)
	m.bounds = bounds
	m.pixels = picture
	return m
}

func NewFromGray(in *image.Gray) *M {
	bounds := in.Bounds()
	m := New(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			m.Set(x, y, float64(in.GrayAt(x, y).Y))
		}
	}
	return m
}

func (m *M) At(x, y int) float64 {
	return m.pixels[y][x]
}

func (m *M) Set(x, y int, value float64) {
	m.pixels[y][x] = value
}

func (m *M) Bounds() image.Rectangle {
	return m.bounds
}

func (m *M) SubImage(r image.Rectangle) *M {
	r = r.Intersect(m.bounds)
	if r.Empty() {
		log.Fatal("empty intersection of bounds")
	}
	return &M{
		pixels: m.pixels,
		bounds: r,
	}
}

func (m *M) ToGray() *image.Gray {
	bounds := m.Bounds()
	gray := image.NewGray(bounds)
	var min = 1000.0
	var max float64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := m.At(x, y)
			if pixel < min {
				min = pixel
			}
			if pixel > max {
				max = pixel
			}
		}
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			col := uint8((m.At(x, y) - min) / (max - min) * 255)
			gray.SetGray(x, y, color.Gray{Y: uint8(col)})
		}
	}
	return gray
}

func (m *M) SaveAsJPG(fname string) {
	gray := m.ToGray()
	f, _ := os.Create(fname)
	defer f.Close()
	jpeg.Encode(f, gray, nil)
}
