package named

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

var GeometryFilter = dfl.MustParseCompile("filter(@, '(@geometry?.coordinates != null) and (@geometry.coordinates[0] within $bbox[0] and $bbox[2]) and (@geometry.coordinates[1] within $bbox[1] and $bbox[3])')")

var Length = dfl.MustParseCompile("len(@)")

var Limit = dfl.MustParseCompile("limit(@, $limit)")

var GeoJSONLinesToGeoJSON = dfl.MustParseCompile("map(@, '@properties -= {`_tile_x`, `_tile_y`, `_tile_z`}') | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}")

var GroupByTile = dfl.MustParseCompile("group(@, '[tileY(@geometry.coordinates[1], $z), tileX(@geometry.coordinates[0], $z)]')")
