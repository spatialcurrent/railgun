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

type ErrInvalidParameter struct {
	Name  string
	Value interface{}
}

func (e *ErrInvalidParameter) Error() string {
	return "invalid parameter " + e.Name + " with value " + fmt.Sprint(e.Value)
}
