// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package dfl

import (
	"github.com/spatialcurrent/pflag"
)

// InitDflFlags initializes the dfl flags.
func InitDflFlags(flag *pflag.FlagSet) {
	flag.StringP(FlagDflExpression, "d", "", "DFL expression to use")
	flag.String(FlagDflUri, "", "URI to DFL file to use")
	flag.String(FlagDflVars, "", "initial variables to use when evaluating DFL expression")
}
