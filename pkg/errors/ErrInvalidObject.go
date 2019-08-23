// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

type ErrInvalidObject struct {
	Value interface{}
}

func (e *ErrInvalidObject) Error() string {
	str := "invalid object"
	value, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            e.Value,
		Format:            "json",
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
		Pretty:            false,
	})
	if err == nil {
		str += " : " + string(value)
	}
	return str
}
