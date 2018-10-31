package img

import (
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
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

	return &railgunerrors.ErrUnknownImageExtension{Extension: ext}
}
