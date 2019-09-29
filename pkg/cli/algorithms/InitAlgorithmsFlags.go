// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package algorithms

import (
	"github.com/spatialcurrent/pflag"
)

// InitAlgorithmsFlags initializes the algorithms flags.
func InitAlgorithmsFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagFormat, "f", DefaultFormat, "Output Format")
}
