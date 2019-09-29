// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package jwt

import (
	"crypto/rsa"

	"github.com/aws/aws-sdk-go/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
)

// LoadPrivateKey loads a JWT Private Key from the given string or uri (checks string first).
func LoadPrivateKey(str string, uri string, s3Client *s3.S3) (*rsa.PrivateKey, error) {

	if len(str) > 0 {
		k, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(str))
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing RSA private key from string %q", str)
		}
		return k, nil
	}

	r, _, err := grw.ReadFromResource(&grw.ReadFromResourceInput{
		Uri:        uri,
		Alg:        grw.AlgorithmNone,
		Dict:       grw.NoDict,
		BufferSize: grw.DefaultBufferSize,
		S3Client:   s3Client,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error opening private key at uri %q", uri)
	}

	b, err := r.ReadAllAndClose()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading private key at uri %q", uri)
	}

	k, err := jwt.ParseRSAPrivateKeyFromPEM(b)
	if err != nil {
		return nil, errors.Wrapf(err, "error parsing RSA private key at uri %q", uri)
	}

	return k, nil
}
