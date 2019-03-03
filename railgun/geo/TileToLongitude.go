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

func TileToLongitude(x int, z int) float64 {
	return float64(x)/math.Pow(float64(2), float64(z))*360 - 180
}
