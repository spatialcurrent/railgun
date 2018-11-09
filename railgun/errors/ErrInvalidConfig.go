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

type ErrInvalidConfig struct {
	Name  string
	Value interface{}
}

func (e *ErrInvalidConfig) Error() string {
	return "invalid config for " + e.Name + " with value " + fmt.Sprint(e.Value)
}
