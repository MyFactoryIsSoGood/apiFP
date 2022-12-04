package kernel

import (
	"apiFP/template/matrix"
	"math"
)

type directional struct {
	Base
	gx, gy *matrix.M
}

func (k *directional) Apply(_ *matrix.M, x, y int) float64 {
	dx := k.gx.At(x, y)
	dy := k.gy.At(x, y)
	ang := math.Atan2(dy, dx) + math.Pi/2
	return ang
}
