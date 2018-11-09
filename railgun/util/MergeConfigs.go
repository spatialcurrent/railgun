// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"github.com/spf13/viper"
)

// MergeConfigs merges an array of config from the given uris into the Viper config.
func MergeConfigs(v *viper.Viper, configUris []string) {
	for _, configUri := range configUris {
		MergeConfig(v, configUri)
	}
}
