// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package geojson

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeatureUnmarshalJSON(t *testing.T) {
	j := `{"type": "Feature", "id": 123, "properties": {"foo": "bar"}, "geometry_name": "the_geom", geometry": {"type":"Point", "coordinates": [12.49268, 41.89029]}}`
	expected := Feature{
		Id: 123,
		Properties: map[string]interface{}{
			"foo": "bar",
		},
		GeometryName: "the_geom",
		Geometry:     Point([]float64{12.49268, 41.89029}),
	}
	f := Feature{}
	err := json.Unmarshal(j, &f)
	assert.NoError(t, err)
	assert.Equal(t, expected, p)
}
