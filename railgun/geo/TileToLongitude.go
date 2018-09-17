package geo

import (
	"math"
)

func TileToLongitude(x int, z int) float64 {
	return float64(x)/math.Pow(float64(2), float64(z))*360 - 180
}
