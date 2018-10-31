package railgun

import (
	"fmt"
	"github.com/spatialcurrent/railgun/railgun/geo"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"strconv"
)

type Tile struct {
	Z int
	X int
	Y int
}

func (t Tile) String() string {
	return fmt.Sprint(t.Z) + "/" + fmt.Sprint(t.X) + "/" + fmt.Sprint(t.Y)
}

func (t Tile) Map() map[string]interface{} {
	return map[string]interface{}{
		"z": t.Z,
		"x": t.X,
		"y": t.Y,
	}
}

func (t Tile) Bbox() []float64 {
	return geo.TileToBoundingBox(t.Z, t.X, t.Y)
}

func NewTileFromRequestVars(vars map[string]string) (Tile, error) {
	t := Tile{}

	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		return t, &railgunerrors.ErrInvalidParameter{Name: "z", Value: vars["z"]}
	}
	t.Z = z

	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		return t, &railgunerrors.ErrInvalidParameter{Name: "x", Value: vars["x"]}
	}
	t.X = x

	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		return t, &railgunerrors.ErrInvalidParameter{Name: "y", Value: vars["y"]}
	}
	t.Y = y
	return t, nil
}
