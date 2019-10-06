// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/viper"
	"github.com/spatialcurrent/cobra"
	"strings"
	"github.co/spatialcurrent/railgun/pkg/util"
)

const (
	CliUse   = "catalog"
	CliShort = "catalog commands for interacting with a local Railgun catalog"
	CliLong  = "catalog commands for interacting with a local Railgun catalog"
)

const (
	FlagCatalogUri = "catalog-uri"
	FlagCatalogFormat = "catalog-format"
	FlagCatalogCompression = "catalog-compression"
)

var (
	ErrMissingCatalogUri = errors.New("missing catalog uri")
	ErrMissingCatalogFormat = errors.New("missing catalog format")
)

func stringSliceContains(slc []string, str string) bool {
	for _, x := range slc {
		if x == str {
			return true
		}
	}
	return false
}

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()
	err := v.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, errors.Wrap(err, "error binding flags")
	}
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	util.MergeConfigs(v, v.GetStringArray("config-uri"))
	return v, nil
}