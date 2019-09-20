// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package binary

import (
	"encoding"

	"github.com/pkg/errors"
)

func Unmarshal(b []byte, v interface{}) error {
	if u, ok := v.(encoding.BinaryUnmarshaler); ok {
		return u.UnmarshalBinary(b)
	}
	return errors.Errorf("object %#v (%T) does not implement encoding.BinaryUnmarshaler", v, v)
}
