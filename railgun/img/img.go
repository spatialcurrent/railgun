// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package img

import (
	"image/color"
)

var RedImage = CreateImage(color.RGBA{255, 0, 0, 255})

var GreenImage = CreateImage(color.RGBA{0, 255, 0, 255})

var BlueImage = CreateImage(color.RGBA{0, 0, 255, 255})

var BlankImage = CreateImage(color.RGBA{0, 0, 0, 0})
