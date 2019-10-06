// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package runtime

import (
	"github.com/spatialcurrent/viper"
)

// CheckRuntimeConfig checks the Runtime configuration.
func CheckRuntimeConfig(v *viper.Viper) error {
	procs := v.GetInt(FlagRuntimeMaxProcs)
	if procs < 0 {
		return &ErrInvalidMaxProcs{Value: procs}
	}
	return nil
}
