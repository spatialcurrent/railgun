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

func TestFunctionMap(t *testing.T) {
	f := &Function{
		Name:        "foo",
		Title:       "Foo",
		Description: "Foo bar",
		Aliases:     []string{"foo"},
		Node:        nil,
		Tags:        []string{"foo"},
	}

	expected := map[string]interface{}{
		"name":        "foo",
		"title":       "Foo",
		"Description": "Foo bar",
		"aliases":     []string{"foo"},
		"node":        nil,
		"tags":        []string{"foo"},
	}

	out, err := f.Map()
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}
