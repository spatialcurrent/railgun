// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
)

func DecryptReader(reader grw.ByteReadCloser, passphrase string, salt string) ([]byte, error) {
	encrypted, err := reader.ReadAllAndClose()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error reading from resource")
	}

	plain, err := DecryptBytes(encrypted, passphrase, salt)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error decoding input")
	}

	return plain, nil
}