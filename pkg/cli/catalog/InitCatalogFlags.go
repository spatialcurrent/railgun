// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/spatialcurrent/pflag"
)

// InitCatalogFlags initializes the catalog flags.
func InitCatalogFlags(flag *pflag.FlagSet) {
	flag.String(FlagCatalogUri, "", "uri to the catalog file")
	flag.String(FlagCatalogFormat, "", "the format of the catalog file")
	flag.String(FlagCatalogCompression, "", "the compression algorithm of the catalog file")
}
