// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geo

import (
	"math"
)

func LatitudeToTile(lat float64, z int) int {
	lat_rad := lat * math.Pi / 180.0
	return int((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * math.Pow(float64(2), float64(z)))
}
