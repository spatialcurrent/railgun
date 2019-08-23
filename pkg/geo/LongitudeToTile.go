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

func LongitudeToTile(lon float64, z int) int {
	return int((180 + lon) * (math.Pow(float64(2), float64(z)) / 360.0))
}
