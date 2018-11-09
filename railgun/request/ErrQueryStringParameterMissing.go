// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

type ErrQueryStringParameterMissing struct {
	Name string
}

func (e ErrQueryStringParameterMissing) Error() string {
	return "query string parameter " + e.Name + " is missing"
}
