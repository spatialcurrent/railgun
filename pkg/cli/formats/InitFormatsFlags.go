// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package formats

import (
	"github.com/spatialcurrent/pflag"
)

// InitFormatsFlags initializes the formats flags.
func InitFormatsFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagFormat, "f", DefaultFormat, "Output Format")
}