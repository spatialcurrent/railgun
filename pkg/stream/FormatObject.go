// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

// FormatObject formats an object for writing to a stream given an output format.
// If the output format is jsonl, then will format the object as JSON.
// If the output format is csv or tsv, will format the object without a header line.
func FormatObject(object interface{}, format string, header []interface{}, lineSeparator string) ([]byte, error) {
	if format == "jsonl" {
		b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
			Object:            object,
			Format:            "json",
			Header:            header,
			Limit:             gss.NoLimit,
			Pretty:            false,
			LineSeparator:     lineSeparator,
			KeyValueSeparator: "=",
		})
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "error serializing object")
		}
		return b, nil
	}
	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            object,
		Format:            format,
		Header:            header,
		Limit:             gss.NoLimit,
		Pretty:            false,
		LineSeparator:     lineSeparator,
		KeyValueSeparator: "=",
	})
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error serializing object")
	}
	return b, nil
}
