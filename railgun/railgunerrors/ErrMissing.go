// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgunerrors

type ErrMissing struct {
	Type string
	Name string
}

func (e *ErrMissing) Error() string {
	return e.Type + " with name " + e.Name + " is missing"
}
