// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package jwt

import (
	"github.com/spatialcurrent/viper"
)

// CheckJwtConfig checks the JWT configuration.
func CheckJwtConfig(v *viper.Viper) error {
	sessionDuration := v.GetDuration(FlagJwtSessionDuration)
	if sessionDuration.Nanoseconds() < MinJwtSessionDuration.Nanoseconds() {
		return &ErrInvalidSessionDuration{Value: sessionDuration, Min: MinJwtSessionDuration}
	}
	publicKeyString := v.GetString(FlagJwtPublicKey)
	publicKeyUri := v.GetString(FlagJwtPublicKeyUri)
	if len(publicKeyString) == 0 && len(publicKeyUri) == 0 {
		return ErrMissingJwtPublicKey
	}
	privateKeyString := v.GetString(FlagJwtPrivateKey)
	privateKeyUri := v.GetString(FlagJwtPrivateKeyUri)
	if len(privateKeyString) == 0 && len(privateKeyUri) == 0 {
		return ErrMissingJwtPrivateKey
	}
	validMethods := v.GetStringSlice(FlagJwtValidMethods)
	if len(validMethods) == 0 {
		return ErrMissingJwtValidMethods
	}
	return nil
}
