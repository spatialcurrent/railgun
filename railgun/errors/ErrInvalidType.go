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
	str := "invalid type " + fmt.Sprint(reflect.TypeOf(e.Value))
	if reflect.ValueOf(e.Type).IsValid() {
		str += ", expecting " + fmt.Sprint(e.Type)
	}
	return str
}
