// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

// CreateCipher returns a cipher.Block given a salt and passphrase
func CreateCipher(salt_string string, passphrase_string string) (cipher.Block, error) {
	salt_bytes := []byte(salt_string)
	salt := make([]byte, hex.DecodedLen(len(salt_bytes)))
	_, err := hex.Decode(salt, salt_bytes)
	if err != nil {
		return nil, errors.Wrap(err, "invalid salt "+salt_string)
	}
	key := argon2.Key([]byte(passphrase_string), salt, 3, 32*1024, 4, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return block, errors.Wrap(err, "error creating new AES256 cipher")
	}
	return block, nil
}
