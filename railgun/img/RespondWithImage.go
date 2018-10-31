package img

import (
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
)

func RespondWithImage(ext string, w http.ResponseWriter, img *image.RGBA) error {
	switch ext {
	case "gif":
		return gif.Encode(w, img, nil)
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, nil)
	case "png":
		return png.Encode(w, img)
	}
	return &railgunerrors.ErrUnknownImageExtension{Extension: ext}
}
