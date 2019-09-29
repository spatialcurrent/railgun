// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"reflect"

	"github.com/spatialcurrent/go-simple-serializer/pkg/bson"
	"github.com/spatialcurrent/go-simple-serializer/pkg/serializer"
	"github.com/spatialcurrent/go-simple-serializer/pkg/sv"
	"github.com/spatialcurrent/go-simple-serializer/pkg/tags"
	"github.com/spatialcurrent/go-simple-serializer/pkg/toml"
	"github.com/spatialcurrent/railgun/pkg/geojson"
)

// InitInputType initializes the input type based on the provided type name "t" and format "f".
// Returns error if typ name "t" is set, but known.
func InitInputType(t string, f string) (reflect.Type, error) {
	if len(t) > 0 {
		switch t {
		case "geojson.Feature":
			return reflect.SliceOf(geojson.TypeFeature), nil
		case "geojson.Geometry":
			return reflect.SliceOf(geojson.TypeGeometry), nil
		case "geojson.Point":
			return reflect.SliceOf(geojson.TypePoint), nil
		case "map[string]string":
			return reflect.TypeOf([]map[string]string{}), nil
		case "map[string]interface{}", "map[string]interface {}", "map[string]interface":
			return reflect.TypeOf([]map[string]interface{}{}), nil
		case "interface{}", "interface":
			return reflect.TypeOf([]interface{}{}), nil
		}
		return nil, &ErrUnknownInputType{Value: t}
	}

	switch f {
	case serializer.FormatBSON:
		return bson.DefaultType, nil
	case serializer.FormatCSV, serializer.FormatTSV:
		return reflect.SliceOf(sv.DefaultType), nil
	case serializer.FormatTags:
		return reflect.SliceOf(tags.DefaultType), nil
	case serializer.FormatTOML:
		return toml.DefaultType, nil
	}

	return nil, nil
}
