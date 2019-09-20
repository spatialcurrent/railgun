// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package client

import (
	"github.com/spatialcurrent/pflag"
	"github.com/spatialcurrent/railgun/pkg/cli/output"
)

// InitClientFlags initializes the client flags.
func InitClientFlags(flag *pflag.FlagSet) {
	flag.String(FlagJwtToken, "", "The JWT token")
	flag.StringP(FlagServer, "s", DefaultServer, "the location of the server")
	output.InitOutputFlags(flag, "json")
}
