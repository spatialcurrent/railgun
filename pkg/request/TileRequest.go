// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

import (
	"fmt"
	"net/http"

	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/railgun/pkg/core"
)

type TileRequest struct {
	Header        http.Header
	Layer         string
	Tile          core.Tile
	Bbox          []float64
	Source        string
	Expression    string
	Features      int
	OutsideExtent bool
}

func (tr TileRequest) String() string {
	str := "requested tile " + tr.Tile.String() + " for layer " + tr.Layer + " from " + tr.Source
	if len(tr.Expression) > 0 {
		str += " with filter " + tr.Expression
	}
	str += " has " + fmt.Sprint(tr.Features) + " features within bounding box " + fmt.Sprint(tr.Bbox)
	return str
}

func (tr TileRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"layer": map[string]interface{}{
			"name": tr.Layer,
		},
		"bbox":          tr.Bbox,
		"outsideExtent": tr.OutsideExtent,
		"tile":          tr.Tile.Map(),
		"source": map[string]interface{}{
			"uri": tr.Source,
		},
		"http": map[string]interface{}{
			"header": tr.Header,
		},
		"dfl": map[string]interface{}{
			"expression": tr.Expression,
		},
		"results": map[string]interface{}{
			"features": tr.Features,
		},
	}
}

func (tr TileRequest) Serialize(format string) ([]byte, error) {
	return gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            tr.Map(),
		Format:            format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
}