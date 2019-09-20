// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package client

import (
	"github.com/spatialcurrent/pflag"
)

// InitClientAuthFlags initializes the client auth flags.
func InitClientAuthFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagUser, "u", "", "user")
	flag.StringP(FlagPassword, "p", "", "password")
}
