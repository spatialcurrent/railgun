// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/spatialcurrent/viper"
)

// CheckCatalogConfig checks the catalog configuration.
func CheckCatalogConfig(v *viper.Viper, formats string, algorithms []string) error {
	uri := v.GetString(FlagCatalogUri)
	if len(uri) == 0 {
		return ErrMissingCatalogUri
	}
	catalogFormat := v.GetString(FlagCatalogFormat)
	if len(catalogFormat) == 0 {
		return &ErrMissingCatalogFormat{Expected: formats}
	}
	if !stringSliceContains(formats, catalogFormat) {
		return &ErrInvalidCatalogFormat{Value: catalogFormat, Expected: formats}
	}
	catalogCompression := v.GetString(FlagCatalogCompression)
	if len(catalogCompression) > 0 && !stringSliceContains(algorithms, catalogCompression) {
		return &ErrInvalidCatalogCompression{Value: catalogCompression, Expected: algorithms}
	}
	return nil
}
