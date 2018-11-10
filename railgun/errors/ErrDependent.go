// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

type ErrDependent struct {
	Type          string
	Name          string
	DependentType string
	DependentName string
}

func (e *ErrDependent) Error() string {
	return "existing " + e.DependentType + " with name " + e.DependentName + " is dependent on " + e.Type + " with name " + e.Name
}
