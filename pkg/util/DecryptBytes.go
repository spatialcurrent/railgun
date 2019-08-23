// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

func DecryptBytes(input []byte, passphrase string, salt string) ([]byte, error) {

	if len(passphrase) > 0 {
		block, err := CreateCipher(salt, passphrase)
		if err != nil {
			return make([]byte, 0), errors.New("error creating cipher for input passphrase")
		}
		ciphertext := input
		if len(ciphertext) < aes.BlockSize {
			return make([]byte, 0), errors.New("cipher text is too short: cipher text is shorter than the AES block size.")
		}
		iv := ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]
		stream := cipher.NewCFBDecrypter(block, iv)
		stream.XORKeyStream(ciphertext, ciphertext)
		return ciphertext, nil
	}

	return input, nil
}
