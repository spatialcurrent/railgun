// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geojson

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/binary"
)

type FeatureId int64

type Feature struct {
	Id           FeatureId              `json:"id" yaml:"id"`
	Properties   map[string]interface{} `json:"properties" yaml:"properties`
	GeometryName string                 `json:"geometry_name" yaml:"geometry_name"`
	Geometry     Geometry               `json:"geometry" yaml:"geometry"`
}

func (f *Feature) Type() string {
	return TypeNameFeature
}

func (f *Feature) UnmarshalJSON(b []byte) error {
	s := struct {
		Id           FeatureId              `json:"id" yaml:"id"`
		Properties   map[string]interface{} `json:"properties" yaml:"properties`
		GeometryName string                 `json:"geometry_name" yaml:"geometry_name"`
		Geometry     json.RawMessage        `json:"coordinates"`
	}{}

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	f.Id = s.Id
	f.Properties = s.Properties
	f.GeometryName = s.GeometryName

	if len(s.Geometry) > 0 {
		g := struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		}{}
		err = json.Unmarshal(s.Geometry, &g)
		if err != nil {
			return err
		}
		switch g.Type {
		case TypeNamePoint:
			coordinates := make([]float64, 0)
			err := json.Unmarshal(g.Coordinates, &coordinates)
			if err != nil {
				return err
			}
			p := Point(coordinates)
			f.Geometry = &p
		}
	}

	return nil
}

func (f Feature) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":            f.Id,
		"type":          f.Type(),
		"properties":    f.Properties,
		"geometry_name": f.GeometryName,
		"geometry":      f.Geometry,
	})
}

func (f Feature) GobEncode() ([]byte, error) {

	buf := new(bytes.Buffer)

	err := buf.WriteByte(FlagFeature)
	if err != nil {
		return buf.Bytes(), errors.Wrap(err, "error writing feature flag")
	}

	if f.Geometry == nil {
		err = buf.WriteByte(FlagNone)
		if err != nil {
			return buf.Bytes(), errors.Wrap(err, "error writing none flag")
		}
	} else {
		g, err := binary.Marshal(f.Geometry)
		if err != nil {
			return buf.Bytes(), errors.Wrap(err, "error marshalling geometry to binary")
		}

		_, err = buf.Write(g)
		if err != nil {
			return buf.Bytes(), errors.Wrap(err, "error writing geometry to buffer")
		}
	}

	enc := gob.NewEncoder(buf)

	err = enc.Encode(f.Id)
	if err != nil {
		return buf.Bytes(), errors.Wrapf(err, "error encoding id %v", f.Id)
	}

	err = enc.Encode(f.GeometryName)
	if err != nil {
		return buf.Bytes(), errors.Wrapf(err, "error encoding geometry name %q", f.GeometryName)
	}

	err = enc.Encode(f.Properties)
	if err != nil {
		return buf.Bytes(), errors.Wrapf(err, "error encoding properties %v", f.Properties)
	}

	return buf.Bytes(), nil
}

func (f *Feature) GobDecode(data []byte) error {

	if len(data) < 2 {
		return fmt.Errorf("invalid byte stream for feature, length is %d", len(data))
	}

	i := 0

	if data[i] != FlagFeature {
		return fmt.Errorf("invalid flag for feature %d, expecting %d", data[0], FlagFeature)
	}

	i++

	switch data[i] {
	case FlagNone:
	case FlagPoint:
		p := Point([]float64{})
		err := binary.Unmarshal(data[i:i+17], &p)
		if err != nil {
			return errors.Wrap(err, "error unmarshaling point bytes")
		}
		f.Geometry = &p
		i += 17
	default:
		return fmt.Errorf("invalid flag for geometry %d", data[1])
	}

	r := bytes.NewReader(data[i:])

	dec := gob.NewDecoder(r)

	err := dec.Decode(&f.Id)
	if err != nil {
		return errors.Wrap(err, "error decoding feature id")
	}

	err = dec.Decode(&f.GeometryName)
	if err != nil {
		return errors.Wrap(err, "error decoding geometry name")
	}

	err = dec.Decode(&f.Properties)
	if err != nil {
		return errors.Wrap(err, "error decoding properties")
	}
	return nil
}
