// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cors

import (
	"github.com/spatialcurrent/pflag"
)

// InitCorsFlags initializes the CORS flags.
func InitCorsFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagCorsOrigin, "", CorsOriginWildcard, "value for Access-Control-Allow-Origin header")
	flag.StringP(FlagCorsCredentials, "", "false", "value for Access-Control-Allow-Credentials header")
}
