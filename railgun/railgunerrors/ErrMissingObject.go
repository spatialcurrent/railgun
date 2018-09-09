// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgunerrors

type ErrMissingObject struct {
	Type string
	Name string
}

func (e *ErrMissingObject) Error() string {
	return e.Type + " with name " + e.Name + " does not exist or otherwise missing"
}
