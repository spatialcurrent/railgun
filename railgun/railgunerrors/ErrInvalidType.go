// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgunerrors

import (
	"fmt"
	"reflect"
)

type ErrInvalidType struct {
	Type  reflect.Type
	Value interface{}
}

func (e *ErrInvalidType) Error() string {
	return "invalid type " + fmt.Sprint(reflect.TypeOf(e.Value)) + ", expecting " + fmt.Sprint(e.Type)
}
