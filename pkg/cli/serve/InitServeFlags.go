// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serve

import (
	"github.com/spatialcurrent/pflag"
	"github.com/spatialcurrent/railgun/pkg/cli/cors"
	"github.com/spatialcurrent/railgun/pkg/cli/http"
	"github.com/spatialcurrent/railgun/pkg/cli/jwt"
	"github.com/spatialcurrent/railgun/pkg/cli/swagger"
)

// InitServeFlags initializes the serve flags.
func InitServeFlags(flag *pflag.FlagSet) {
	flag.StringArrayP("datastore", "d", []string{}, "datastore")
	flag.StringArrayP("workspace", "w", []string{}, "workspace")
	flag.StringArrayP("layer", "l", []string{}, "layer")
	flag.StringArrayP("process", "p", []string{}, "process")
	flag.StringArrayP("service", "s", []string{}, "service")
	flag.StringArrayP("job", "j", []string{}, "job")

	http.InitHttpFlags(flag)

	// Cache Flags
	flag.DurationP(FlagCacheDefaultExpiration, "", DefaultCacheDefaultExpiration, "the default expiration for items in the cache")
	flag.DurationP(FlagCacheCleanupInterval, "", DefaultCacheCleanupInterval, "the cleanup interval for the cache")

	// Input Flags
	flag.StringP(FlagInputPassphrase, "", "", "input passphrase for AES-256 encryption")
	flag.StringP(FlagInputSalt, "", "", "input salt for AES-256 encryption")
	flag.IntP(FlagInputReaderBufferSize, "", DefaultInputReaderBufferSize, "the buffer size for the input reader")

	// Logging Flags
	flag.BoolP(FlagLogRequestsTile, "", false, "log tile requests")
	flag.BoolP(FlagLogRequestsCache, "", false, "log cache hit/miss")

	// Mask Flags
	flag.IntP(FlagMaskMinZoom, "", DefaultMaskMinZoom, "minimum mask zoom leel")
	flag.IntP(FlagMaskMaxZoom, "", DefaultMaskMaxZoom, "maximum mask zoom level")

	swagger.InitSwaggerFlags(flag)

	// CORS Flags
	cors.InitCorsFlags(flag)

	// Catalog Skip Errors
	flag.String(FlagCatalogUri, "", "uri of the catalog backend")
	flag.BoolP(FlagConfigSkipErrors, "", false, "skip loading config with bad errors")

	// Security
	flag.String(FlagRootPassword, "", "root user password")

	jwt.InitJwtFlags(flag)

	// Tile
	flag.IntP(FlagTileRandomDelay, "", DefaultTileRandomDelay, "random delay for processing tiles in milliseconds")

	// Coconut
	flag.String(FlagCoconutApiUrl, DefaultCoconutApiUrl, "The API_URL for Coconut.")
	flag.String(FlagCoconutBaselayerUrl, DefaultCoconutBaselayerUrl, "The BASELAYER_URL for Coconut.")
	flag.String(FlagCoconutBundleUrl, DefaultCoconutBundleUrl, "The url to the bundle.js for Coconut.")
}
