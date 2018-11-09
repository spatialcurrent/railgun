// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================
package img

import (
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
)

func RespondWithGrid(ext string, writer http.ResponseWriter, grid []uint8, width int, height int, f color.RGBA, b color.RGBA) error {

	i := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[(y*width)+x] > 0 {
				i.Set(x, y, f)
			} else {
				i.Set(x, y, b)
			}
		}
	}

	switch ext {
	case "gif":
		return gif.Encode(writer, i, nil)
	case "jpeg", "jpg":
		return jpeg.Encode(writer, i, nil)
	case "png":
		return png.Encode(writer, i)
	}

	return &rerrors.ErrUnknownImageExtension{Extension: ext}
}
