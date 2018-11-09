// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

type Request interface {
	String() string
	Map() map[string]interface{}
	Serialize(format string) (string, error)
}
