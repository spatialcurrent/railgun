// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

import (
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

type CacheRequest struct {
	Key string
	Hit bool
}

func (cr CacheRequest) String() string {
	str := "cache"
	if cr.Hit {
		str += " hit"
	} else {
		str += " miss"
	}
	str += " for key " + cr.Key
	return str
}

func (cr CacheRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"key": cr.Key,
		"hit": cr.Hit,
	}
}

func (cr CacheRequest) Serialize(format string) ([]byte, error) {
	return gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            cr.Map(),
		Format:            format,
		Header:            gss.NoHeader,
		Limit:             gss.NoLimit,
		LineSeparator:     "\n",
		KeyValueSeparator: "=",
	})
}
