// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================
package railgunerrors

type ErrMissingRequiredParameter struct {
	Name string
}

func (e *ErrMissingRequiredParameter) Error() string {
	return "required parameter with name " + e.Name + " is missing"
}
