// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

type ErrMissingMethod struct {
	Type   string
	Method string
}

func (e *ErrMissingMethod) Error() string {
	return e.Type + " is missing method " + e.Method
}
