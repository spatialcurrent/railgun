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

func Marshal(v interface{}) ([]byte, error) {
	if u, ok := v.(encoding.BinaryMarshaler); ok {
		return u.MarshalBinary()
	}
	return make([]byte, 0), errors.Errorf("object %#v (%T) does not implement encoding.BinaryUnmarshaler", v, v)
}
