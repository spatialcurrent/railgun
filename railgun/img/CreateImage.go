package img

import (
	"image"
	"image/color"
	"image/draw"
)

func CreateImage(c color.RGBA) *image.RGBA {
	i := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.Draw(i, i.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)
	return i
}
