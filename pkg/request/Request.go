// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

type Request interface {
	String() string
	Map() map[string]interface{}
	Serialize(format string) ([]byte, error)
}
