// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geojson

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Point []float64

func (p Point) Type() string {
	return TypeNamePoint
}

func (p *Point) UnmarshalJSON(b []byte) error {
	s := struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}{}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*p = s.Coordinates
	return nil
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":        TypePoint,
		"coordinates": p,
	})
}

func (p Point) MarshalBinary() ([]byte, error) {
	if len(p) == 2 {
		return make([]byte, 0), fmt.Errorf("invalid number of elements in point slice (%d), expecting 2", len(p))
	}

	buf := new(bytes.Buffer)

	err := buf.WriteByte(FlagPoint)
	if err != nil {
		return buf.Bytes(), errors.Wrap(err, "error writing point flag")
	}

	for i := 0; i < 2; i++ {
		err := binary.Write(buf, binary.LittleEndian, p[i])
		if err != nil {
			return buf.Bytes(), errors.Wrapf(err, "error writing float %d for point", i)
		}
	}

	return buf.Bytes(), nil
}

func (p *Point) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("invalid byte stream for point, length is %d, expecting 17", len(data))
	}
	if data[0] != FlagPoint {
		return fmt.Errorf("invalid flag for point %d, expecting %d", data[0], FlagPoint)
	}
	r := bytes.NewReader(data[1:])
	coordinates := make([]float64, 0)
	for i := 0; i < 2; i++ {
		c := 0.0
		err := binary.Read(r, binary.LittleEndian, &c)
		if err != nil {
			return errors.Wrapf(err, "error reading float %d for point", i)
		}
		coordinates = append(coordinates, c)
	}
	*p = coordinates
	return nil
}
