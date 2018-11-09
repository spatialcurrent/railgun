// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgunerrors

import (
	"fmt"
)

type ErrAlreadyExists struct {
	Name  string
	Value interface{}
}

func (e *ErrAlreadyExists) Error() string {
	return "config " + e.Name + " with name " + fmt.Sprint(e.Value) + " already exists"
}
