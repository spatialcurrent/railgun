// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceMap(t *testing.T) {
	w := &Workspace{
		Name:        "foo",
		Title:       "Foo",
		Description: "Foo bar",
	}

	out, err := W.Map()
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"name": "foo", "title": "Foo", "Description": "Foo bar"}, out)
}
