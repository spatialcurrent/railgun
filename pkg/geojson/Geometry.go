// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geojson

import (
	"encoding"
	"encoding/json"
	//"github.com/go-spatial/geom/encoding/geojson"
	//"github.com/go-spatial/geom/encoding/wkb"
)

type Geometry interface {
	Type() string
	json.Marshaler
	json.Unmarshaler
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

/*

func (g *Geometry) GobDecode(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("unexpected number of bytes (%v) for %T, expecting 1", len(b), g)
	}

	switch uint8(b[0]) {
	case GeometryTypePoint:
		g.Type = "Point"
	case GeometryTypePolygon:
		g.Type = "Polygon"
	default:
		return fmt.Errorf("unexpected geometry type %d", b[0])
	}

	d := gob.NewDecoder(bytes.NewReader(b[1:]))

	coordinates := make([]interface{}, 0)

	err := d.Decode(&coordinates)
	if err != nil {
		return err
	}

	g.Coordinates = coordinates

	return nil
}

func (g *Geometry) GobEncode() ([]byte, error) {
	b := new(bytes.Buffer)

	switch g.Type {
	case "Point":
		b.WriteByte(GeometryTypePoint)
	case "Polygon":
		b.WriteByte(GeometryTypePolygon)
	default:
		return make([]byte, 0), fmt.Errorf("unexpected geometry type %v", g.Type)
	}

	e := gob.NewEncoder(b)

	err := e.Encode(g.Coordinates)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (g *Geometry) UnmarshalJSON(b []byte) error {

	switch g.Type {
	case "Point":
		if p, ok := g.Coordinates.([]interface{}); ok && len(p) == 2 {
			return json.Marshal(map[string]interface{}{
				"type":        "Point",
				"coordinates": json.RawMessage([]byte(fmt.Sprintf("[%.5f, %.5f]", p[0], p[1]))),
			})
		}
	}
	fmt.Println(fmt.Sprintf("Unknown coordinates type %v - %T", g.Type, g.Coordinates))
	return json.Marshal(g)
}


func (g *Geometry) MarshalJSON() ([]byte, error) {
	switch g.Type {
	case "Point":
		if p, ok := g.Coordinates.([]interface{}); ok && len(p) == 2 {
			return json.Marshal(map[string]interface{}{
				"type":        "Point",
				"coordinates": json.RawMessage([]byte(fmt.Sprintf("[%.5f, %.5f]", p[0], p[1]))),
			})
		}
	}
	fmt.Println(fmt.Sprintf("Unknown coordinates type %v - %T", g.Type, g.Coordinates))
	return json.Marshal(g)
}

*/
