// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanStream(t *testing.T) {

	/* CSV */
	assert.True(t, CanStream("csv", "csv", false))
	assert.False(t, CanStream("csv", "json", false))
	assert.True(t, CanStream("csv", "jsonl", false))
	assert.True(t, CanStream("csv", "go", false))
	assert.True(t, CanStream("csv", "gob", false))
	assert.True(t, CanStream("csv", "tsv", false))

	/* JSONL */
	assert.False(t, CanStream("jsonl", "csv", false))
	assert.False(t, CanStream("jsonl", "json", false))
	assert.True(t, CanStream("jsonl", "jsonl", false))
	assert.True(t, CanStream("jsonl", "go", false))
	assert.True(t, CanStream("jsonl", "gob", false))
	assert.False(t, CanStream("jsonl", "tsv", false))

	/* JSONL */
	assert.False(t, CanStream("gob", "csv", false))
	assert.False(t, CanStream("gob", "json", false))
	assert.True(t, CanStream("gob", "jsonl", false))
	assert.True(t, CanStream("gob", "go", false))
	assert.True(t, CanStream("gob", "gob", false))
	assert.False(t, CanStream("gob", "tsv", false))

}
