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

func TestPointUnmarshalJSON(t *testing.T) {
	j := `{"type":"Point", "coordinates": [12.49268, 41.89029]}`
	p := Point([]float64{})
	err := json.Unmarshal(j, &p)
	assert.NoError(t, err)
	assert.Equal(t, Point([]float64{12.49268, 41.89029}), p)
}
