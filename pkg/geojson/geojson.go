// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geojson

import (
	"reflect"
)

const (
	TypeNameFeature = "Feature"
	TypeNamePoint   = "Point"

	FlagNone    = byte(0)
	FlagFeature = byte(1)
	FlagPoint   = byte(2)
)

var (
	TypeFeature  = reflect.TypeOf(Feature{})
	TypePoint    = reflect.TypeOf(Point{})
	TypeGeometry = reflect.TypeOf(Geometry(nil))
)
