// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"fmt"
	"strconv"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/geo"
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

func (t Tile) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range t.Map() {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

func (t Tile) Bbox() []float64 {
	return geo.TileToBoundingBox(t.Z, t.X, t.Y)
}

func NewTileFromRequestVars(vars map[string]string) (Tile, error) {
	t := Tile{}

	z, err := strconv.Atoi(vars["z"])
	if err != nil {
		return t, &rerrors.ErrInvalidParameter{Name: "z", Value: vars["z"]}
	}
	t.Z = z

	x, err := strconv.Atoi(vars["x"])
	if err != nil {
		return t, &rerrors.ErrInvalidParameter{Name: "x", Value: vars["x"]}
	}
	t.X = x

	y, err := strconv.Atoi(vars["y"])
	if err != nil {
		return t, &rerrors.ErrInvalidParameter{Name: "y", Value: vars["y"]}
	}
	t.Y = y
	return t, nil
}
