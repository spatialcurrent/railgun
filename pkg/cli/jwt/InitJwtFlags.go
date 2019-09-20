// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package jwt

import (
	"github.com/spatialcurrent/pflag"
)

// InitJwtFlags initializes the JWT flags.
func InitJwtFlags(flag *pflag.FlagSet) {
	flag.String(FlagJwtPrivateKey, "", "Private RSA Key for JWT")
	flag.String(FlagJwtPrivateKeyUri, "", "URI to private RSA Key for JWT")
	flag.String(FlagJwtPublicKey, "", "Public RSA Key for JWT")
	flag.String(FlagJwtPublicKeyUri, "", "URI to public RSA Key for JWT")
	flag.StringArray(FlagJwtValidMethods, DefaultJwtValidMethods, "Valid methods for JWT")
	flag.Duration(FlagJwtSessionDuration, DefaultJwtSessionDuration, "duration of authenticated session")
}
