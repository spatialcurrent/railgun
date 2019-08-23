// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

type ErrMissingObject struct {
	Type string
	Name string
}

func (e *ErrMissingObject) Error() string {
	return e.Type + " with name " + e.Name + " does not exist or otherwise missing"
}
