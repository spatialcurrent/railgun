// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package jwt

import (
	"time"

	"github.com/pkg/errors"
)

const (
	FlagJwtPrivateKey         = "jwt-private-key"
	FlagJwtPrivateKeyUri      = "jwt-private-key-uri"
	FlagJwtPublicKey          = "jwt-public-key"
	FlagJwtPublicKeyUri       = "jwt-public-key-uri"
	FlagJwtValidMethods       = "jwt-valid-metods"
	FlagJwtSessionDuration    = "jwt-session-duration"
	DefaultJwtSessionDuration = 60 * time.Minute

	MinJwtSessionDuration = time.Second * 30
)

var (
	DefaultJwtValidMethods = []string{"RS512"}
)

var (
	ErrMissingJwtPrivateKey   = errors.New("missing JWT private key, either jwt-private-key or jwt-private-key-uri is required")
	ErrMissingJwtPublicKey    = errors.New("missing JWT public key, either jwt-public-key or jwt-public-key-uri is required")
	ErrMissingJwtValidMethods = errors.New("missing JWT valid methods, at least one is required")
)
