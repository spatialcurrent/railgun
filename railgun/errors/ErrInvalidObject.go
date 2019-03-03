// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

import (
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

type ErrInvalidObject struct {
	Value interface{}
}

func (e *ErrInvalidObject) Error() string {
	str := "invalid object"
	value, err := gss.SerializeString(e.Value, "json", gss.NoHeader, gss.NoLimit)
	if err == nil {
		str += " : " + value
	}
	return str
}
