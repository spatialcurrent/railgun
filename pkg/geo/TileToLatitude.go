// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geo

import (
	"math"
)

var R2D = 180 / math.Pi

func TileToLatitude(y int, z int) float64 {
	n := math.Pi - 2*math.Pi*float64(y)/math.Pow(float64(2), float64(z))
	return (R2D * math.Atan(0.5*(math.Exp(n)-math.Exp(-1.0*n))))
}
