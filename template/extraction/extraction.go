package extraction

import (
	"apiFP/template/kernel"
	"apiFP/template/matrix"
	"apiFP/template/utils"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
	"sync"
)

const (
	Pore MinType = iota
	Bifurcation
	Termination
	Unknown

	ridge  = 0
	any    = 1
	valley = 255
	t      = true
	f      = false
)

type MinType byte

type Mask [3][3]uint8

// Passes checks if pattern of sector matches mask argument
func (m Mask) Passes(sector Mask) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if m[i][j] != sector[i][j] && m[i][j] != 1 {
				return false
			}
		}
	}
	return true
}

type MasksSet []Mask

// MaskSets for all 3 types of morphological operations
var Skeletonization = MasksSet{
	Mask{
		{ridge, ridge, any},
		{any, ridge, valley},
		{valley, valley, valley},
	},
	Mask{
		{any, ridge, ridge},
		{valley, ridge, any},
		{valley, valley, valley},
	},
	Mask{
		{valley, any, ridge},
		{valley, ridge, ridge},
		{valley, valley, any},
	},
	Mask{
		{valley, valley, any},
		{valley, ridge, ridge},
		{valley, any, ridge},
	},
	Mask{
		{valley, valley, valley},
		{valley, ridge, any},
		{any, ridge, ridge},
	},
	Mask{
		{valley, valley, valley},
		{any, ridge, valley},
		{ridge, ridge, any},
	},
	Mask{
		{any, valley, valley},
		{ridge, ridge, valley},
		{ridge, any, valley},
	},
	Mask{
		{ridge, any, valley},
		{ridge, ridge, valley},
		{any, valley, valley},
	},

	Mask{
		{ridge, ridge, ridge},
		{any, ridge, ridge},
		{valley, valley, ridge},
	},
	Mask{
		{any, ridge, ridge},
		{valley, ridge, ridge},
		{valley, ridge, ridge},
	},
	Mask{
		{valley, any, ridge},
		{valley, ridge, ridge},
		{ridge, ridge, ridge},
	},
	Mask{
		{valley, valley, any},
		{ridge, ridge, ridge},
		{ridge, ridge, ridge},
	},
	Mask{
		{ridge, valley, valley},
		{ridge, ridge, any},
		{ridge, ridge, ridge},
	},
	Mask{
		{ridge, ridge, valley},
		{ridge, ridge, valley},
		{ridge, ridge, any},
	},
	Mask{
		{ridge, ridge, ridge},
		{ridge, ridge, valley},
		{ridge, any, valley},
	},
	Mask{
		{ridge, ridge, ridge},
		{ridge, ridge, ridge},
		{any, valley, valley},
	},
}
var Disconnection = MasksSet{
	Mask{
		{any, any, any},
		{valley, ridge, valley},
		{any, any, any},
	},
	Mask{
		{any, valley, any},
		{any, ridge, any},
		{any, valley, any},
	},
}
var Clean = MasksSet{
	Mask{
		{ridge, ridge, ridge},
		{valley, ridge, valley},
		{valley, valley, valley},
	},
	Mask{
		{valley, valley, valley},
		{valley, ridge, valley},
		{ridge, ridge, ridge},
	},
	Mask{
		{valley, valley, ridge},
		{valley, ridge, ridge},
		{valley, valley, ridge},
	},
	Mask{
		{ridge, valley, valley},
		{ridge, ridge, valley},
		{ridge, valley, valley},
	},

	Mask{
		{ridge, ridge, any},
		{ridge, ridge, valley},
		{ridge, valley, valley},
	},
	Mask{
		{any, ridge, ridge},
		{valley, ridge, ridge},
		{valley, valley, ridge},
	},
	Mask{
		{valley, ridge, ridge},
		{valley, ridge, ridge},
		{ridge, ridge, ridge},
	},
	Mask{
		{ridge, ridge, valley},
		{ridge, ridge, valley},
		{ridge, ridge, ridge},
	},
}

// SemiBin removes all pixels with value less than threshold. But if pixel is higher, it stays the same
func SemiBin(img *image.Gray, thresh uint8) *image.Gray {
	out := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if img.GrayAt(x, y).Y > thresh {
				out.SetGray(x, y, color.Gray{Y: 255})
			} else {
				out.SetGray(x, y, img.GrayAt(x, y))
			}
		}
	}
	return out
}

// Binarize returns new binarized image by threshold
func Binarize(in *image.Gray, threshold uint8) *image.Gray {
	var out = image.NewGray(in.Bounds())
	for y := 0; y < in.Bounds().Dy(); y++ {
		for x := 0; x < in.Bounds().Dx(); x++ {
			if in.GrayAt(x, y).Y > threshold {
				out.Set(x, y, color.Gray{Y: 255})
			} else {
				out.Set(x, y, color.Gray{Y: 0})
			}
		}
	}
	return out
}

// OtsuThresholdValue calculates optimal threshold for picture binarization
func OtsuThresholdValue(img *image.Gray) uint8 {
	var histogram [256]uint64
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			histogram[img.GrayAt(x, y).Y]++
		}
	}
	size := img.Bounds().Size()
	totalNumberOfPixels := size.X * size.Y

	var sumHist float64
	for i, bin := range histogram {
		sumHist += float64(uint64(i) * bin)
	}

	var sumBackground float64
	var weightBackground int
	var weightForeground int

	maxVariance := 0.0
	var thresh uint8
	for i, bin := range histogram {
		weightBackground += int(bin)
		if weightBackground == 0 {
			continue
		}
		weightForeground = totalNumberOfPixels - weightBackground
		if weightForeground == 0 {
			break
		}

		sumBackground += float64(uint64(i) * bin)

		meanBackground := float64(sumBackground) / float64(weightBackground)
		meanForeground := (sumHist - sumBackground) / float64(weightForeground)

		variance := float64(weightBackground) * float64(weightForeground) * (meanBackground - meanForeground) * (meanBackground - meanForeground)

		if variance > maxVariance {
			maxVariance = variance
			thresh = uint8(i)
		}
	}
	return thresh
}

// Resize is a util function to resize image to gray object. Uses imaging.Resize and image.Gray
func Resize(in *image.Gray, width int, height int, filt imaging.ResampleFilter) *image.Gray {
	out := imaging.Resize(in, width, height, filt)
	gray := image.NewGray(out.Bounds())
	draw.Draw(gray, out.Bounds(), out, image.Point{}, draw.Src)
	return gray
}

// Directions return directions matrix for image. Uses Sobel operator for generating gradients by Y and by X
// With obtained gradients we can calculate directions matrix
func Directions(img *image.Gray) *matrix.M {
	normalizedMatrix := matrix.NewFromGray(img)
	bounds := normalizedMatrix.Bounds()
	gx, gy := matrix.New(bounds), matrix.New(bounds)
	directions := matrix.New(bounds)
	kernel.SobelDx.ConvoluteParallelized(normalizedMatrix, gx)
	kernel.SobelDy.ConvoluteParallelized(normalizedMatrix, gy)
	kernel.FilteredDirectional(gx, gy, 4).ConvoluteParallelized(directions, directions)
	return directions
}

// processRowsInParallel is a util function to apply morphological operators(Skeletization, disconnection and clean) to image in parallel
func processRowsInParallel(in *image.Gray, from, to int, mask Mask, blacklist *[][2]int, wg *sync.WaitGroup, mtx *sync.Mutex) {
	bounds := in.Bounds()
	for y := from; y < to; y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if in.GrayAt(x, y).Y == 0 {
				sector := Mask{
					{in.GrayAt(x-1, y-1).Y, in.GrayAt(x, y-1).Y, in.GrayAt(x+1, y-1).Y},
					{in.GrayAt(x-1, y).Y, in.GrayAt(x, y).Y, in.GrayAt(x+1, y).Y},
					{in.GrayAt(x-1, y+1).Y, in.GrayAt(x, y+1).Y, in.GrayAt(x+1, y+1).Y}}
				if mask.Passes(sector) {
					mtx.Lock()
					*blacklist = append(*blacklist, [2]int{x, y})
					mtx.Unlock()
				}
			}
		}
	}
	wg.Done()
}

// Morphological returns new image with applied morphological operator. Works in parallel with 10 constant goroutines
// Without parallelization the process slows down 7 times
func Morphological(in *image.Gray, set MasksSet) {
	bounds := in.Bounds()

	for _, mask := range set {
		changes := true
		for changes {
			changes = false
			var blacklist [][2]int
			//for y := 1; y < bounds.Dy()-1; y++ {
			//	for x := 1; x < bounds.Dx()-1; x++ {
			//		if in.GrayAt(x, y).Y != 0 {
			//			continue
			//		}
			//		sector := Mask{
			//			{in.GrayAt(x-1, y-1).Y, in.GrayAt(x, y-1).Y, in.GrayAt(x+1, y-1).Y},
			//			{in.GrayAt(x-1, y).Y, in.GrayAt(x, y).Y, in.GrayAt(x+1, y).Y},
			//			{in.GrayAt(x-1, y+1).Y, in.GrayAt(x, y+1).Y, in.GrayAt(x+1, y+1).Y},
			//		}
			//		if mask.Passes(sector) {
			//			blacklist = append(blacklist, [2]int{x, y})
			//		}
			//	}
			//}
			var wg sync.WaitGroup
			var mutex sync.Mutex
			thrCount := 8
			for i := 0; i < thrCount; i++ {
				wg.Add(1)
				go processRowsInParallel(in, bounds.Dy()/thrCount*i, bounds.Dy()/thrCount*(i+1), mask, &blacklist, &wg, &mutex)
			}
			wg.Wait()
			changes = len(blacklist) > 0
			for _, point := range blacklist {
				in.Set(point[0], point[1], color.Gray{Y: 255})
			}
		}
	}
}

// ClearSkeleton removes garbage from image. Applies window of size [offset] to each pixel and if borders are white
// fills all window by white
func ClearSkeleton(skeleton *image.Gray, offset int) {
	bounds := skeleton.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			flag := false
			for i := 0; i < offset; i++ {
				if skeleton.GrayAt(x+i, y).Y == 0 {
					flag = true
					break
				} else if skeleton.GrayAt(x+i, y+offset-1).Y == 0 {
					flag = true
					break
				}
			}
			for i := 1; i < offset-1; i++ {
				if skeleton.GrayAt(x, y+i).Y == 0 {
					flag = true
					break
				} else if skeleton.GrayAt(x+offset-1, y+i).Y == 0 {
					flag = true
					break
				}
			}
			if !flag {
				for i := 0; i < offset; i++ {
					for j := 0; j < offset; j++ {
						skeleton.SetGray(x+i, y+j, color.Gray{Y: 255})
					}
				}
			}
		}
	}
}

// Minutia provides minutia extraction from skeletonized image
func Minutia(skeleton *image.Gray, filteredDirectional *matrix.M) []utils.Minutiae {
	minutiaes := []utils.Minutiae{}
	bounds := skeleton.Bounds()
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			minutiaeType := matchMinutiaeType(skeleton, x, y)
			if minutiaeType != Unknown && minutiaeType != Termination {
				minutiae := utils.Minutiae{
					XPos:  uint16(x),
					YPos:  uint16(y),
					Angle: uint8(filteredDirectional.At(x, y) / 0.0175 * 255.0 / 360.0),
					Type:  uint16(minutiaeType),
				}
				minutiaes = append(minutiaes, minutiae)
			}
		}
	}
	return minutiaes
}

// matchMinutiaeType returns minutiae type of the given point
func matchMinutiaeType(in *image.Gray, i, j int) MinType { //TODO: refactor
	p0 := in.GrayAt(i-1, j-1).Y == ridge
	p1 := in.GrayAt(i, j-1).Y == ridge
	p2 := in.GrayAt(i+1, j-1).Y == ridge
	p3 := in.GrayAt(i+1, j).Y == ridge
	p4 := in.GrayAt(i+1, j+1).Y == ridge
	p5 := in.GrayAt(i, j+1).Y == ridge
	p6 := in.GrayAt(i-1, j+1).Y == ridge
	p7 := in.GrayAt(i-1, j).Y == ridge
	pc := in.GrayAt(i, j).Y == ridge

	and := func(f0, f1, f2, f7, fc, f3, f6, f5, f4 bool) bool {
		return (pc == fc) && (p0 == f0) && (p1 == f1) && (p2 == f2) && (p3 == f3) && (p4 == f4) && (p5 == f5) && (p6 == f6) && (p7 == f7)
	}

	isPore := and(f, t, f,
		t, f, t,
		f, t, f)

	if isPore {
		return Pore
	}

	isBifurcation := and(t, f, t,
		f, t, f,
		f, f, t) ||
		and(t, f, t,
			f, t, f,
			f, t, f) ||
		and(t, f, t,
			f, t, f,
			t, f, f) ||
		and(t, f, f,
			f, t, t,
			t, f, f) ||
		and(t, f, f,
			f, t, f,
			t, f, t) ||
		and(f, t, f,
			f, t, f,
			t, f, t) ||
		and(f, f, t,
			f, t, f,
			t, f, t) ||
		and(f, f, t,
			t, t, f,
			f, f, t) ||
		and(t, f, t,
			f, t, f,
			f, f, t) ||
		and(f, f, f,
			t, t, t,
			f, t, f) ||
		and(f, t, f,
			f, t, t,
			f, t, f) ||
		and(f, t, f,
			t, t, t,
			f, f, f) ||
		and(f, t, f,
			t, t, f,
			f, t, f) ||
		and(t, f, f,
			f, t, t,
			f, t, f) ||
		and(f, t, f,
			t, t, f,
			f, f, t) ||
		and(f, f, t,
			t, t, f,
			f, t, f) ||
		and(f, t, f,
			f, t, t,
			t, f, f)

	if isBifurcation {
		return Bifurcation
	}

	isTermination := and(t, f, f,
		f, t, f,
		f, f, f) ||
		and(f, t, f,
			f, t, f,
			f, f, f) ||
		and(f, f, t,
			f, t, f,
			f, f, f) ||
		and(f, f, f,
			f, t, t,
			f, f, f) ||
		and(f, f, f,
			f, t, f,
			f, f, t) ||
		and(f, f, f,
			f, t, f,
			f, t, f) ||
		and(f, f, f,
			f, t, f,
			t, f, f) ||
		and(f, f, f,
			t, t, f,
			f, f, f)

	if isTermination {
		return Termination
	}

	return Unknown
}

// MinToTemplate converts minutiae to template
func MinToTemplate(min []utils.Minutiae) utils.Template {
	var template utils.Template
	template.Header.Magic = "FMR"
	template.Header.Version = " 20"
	template.Header.TotalBytes = uint32(30 + len(min))
	template.Header.Height = 400
	template.Header.Width = 400
	template.Header.ResX = 100
	template.Header.ResY = 100
	template.Header.FpCount = 1
	template.Fingerprints = make([]utils.Fingerprint, 1)
	template.Fingerprints[0].Position = 0
	template.Fingerprints[0].ViewOffset = 0
	template.Fingerprints[0].SampleType = 0
	template.Fingerprints[0].FpQuality = 100
	template.Fingerprints[0].MinutiaeCount = uint8(len(min))
	template.Fingerprints[0].Minutiae = min
	template.Fingerprints[0].ExtBytes = 0
	return template
}

func ImgToTemplate(gray *image.Gray) utils.Template {
	gray = SemiBin(gray, 80)
	resized := Resize(gray, 400, 400, imaging.Lanczos)
	binarized := Binarize(resized, OtsuThresholdValue(gray))
	Morphological(binarized, Disconnection)   //disconnection
	Morphological(binarized, Skeletonization) //skeletonization
	ClearSkeleton(binarized, 10)              // clear dots and other garbage
	Morphological(binarized, Clean)           // clean spurs
	directions := Directions(binarized)

	minutiae := Minutia(binarized, directions)
	im := utils.DisplayMinutiae(minutiae, binarized)
	utils.SavePNGImageColored(im, "minutiae.png")

	template := MinToTemplate(minutiae)
	return template
}
