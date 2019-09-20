// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package swagger

import (
	"github.com/spatialcurrent/pflag"
)

// InitSwaggerFlags initializes the swagger flags.
func InitSwaggerFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagSwaggerContactName, "", "", "contact name for swapper document")
	flag.StringP(FlagSwaggerContactEmail, "", "", "contact email for swapper document")
	flag.StringP(FlagSwaggerContactUrl, "", "", "contact url for swapper document")
}
