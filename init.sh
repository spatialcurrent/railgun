#!/bin/bash
go run cmd/railgun/main.go client workspaces add \
--name osm \
--title osm \
--description 'Workspace for OpenStreetMap data'

go run cmd/railgun/main.go client datastores add \
--workspace osm \
--name amenities \
--title Amenities \
--description 'Amenities from OpenStreetMap' \
--uri '~/Downloads/dc_amenities.geojsonl' \
--extent '[-77.5195609,38.8099849,-76.9102596,39.1546259]'

go run cmd/railgun/main.go client datastores add \
--workspace osm \
--name amenities_tiles \
--title Amenities - Tiles \
--description 'Amenities from OpenStreetMap as Tiles' \
--format 'jsonl' \
--compression 'snappy' \
--uri '($tileZ := 10) | (@z <  $tileZ) ? null : "~/Downloads/data/dc_amenities/10-" + int64(mul(@x, pow(2, sub($tileZ, @z)))) + "-" + int64(mul(@y, pow(2, sub($tileZ, @z)))) + ".geojsonl.sz"' \
--extent '[-77.5195609,38.8099849,-76.9102596,39.1546259]'

go run cmd/railgun/main.go client layers add \
--name amenities \
--title Amenities \
--description 'Amenities from OpenStreetMap' \
--datastore amenities_tiles

go run cmd/railgun/main.go client processes add \
--name extent \
--title Extent \
--description 'Get the extent of a list of GeoJSON features' \
--expression 'map(@, "@geometry.coordinates") | bbox(@)'

go run cmd/railgun/main.go client processes add \
--name centroid \
--title Centroid \
--description 'Get the centroid of a list of GeoJSON features' \
--expression 'map(@, "@geometry.coordinates") | bbox(@) | [mean([@[0], @[2]]), mean([@[1], @[3]])]'

go run cmd/railgun/main.go client processes add \
--name unique \
--title Unique Values \
--description 'Create a list of unique values for a propery from a list of GeoJSON features' \
--expression 'map(@, "lookup(@properties, $name)") | array(set(@)) | filter(@, "@ != null")'

go run cmd/railgun/main.go client processes add \
--name hist \
--title Histogram \
--description 'Create a histogram of values for a propery from a list of GeoJSON features' \
--expression 'filter(@, "lookup(@properties, $name) != null") | hist(@, "lower(lookup(@properties, $name))")'

go run cmd/railgun/main.go client processes add \
--name hist_words \
--title Histogram of Words \
--description 'For each lower case value of a property, create a histogram of words from a second property from a list of GeoJSON features' \
--expression 'filter(@, "(lookup(@properties, $first) != null) and (lookup(@properties, $second) != null)") | hist(@, "lower(lookup(@properties, $first))", "set(split(lookup(@properties, $second), ` `))")'

go run cmd/railgun/main.go client processes add \
--name cuisine \
--title Cuisine \
--description 'Filter a list of GeoJSON features by cuisine' \
--expression 'filter(@, "(@properties?.cuisine ilike $cuisine) or ((@properties?.name != null) and ($cuisine in split(@properties.name,` `)))") | (($limit > 0) ? limit(@, $limit) : @) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client services add \
--name amenities_extent \
--title 'Extent of Amenities' \
--description 'Get the extent of amenities' \
--datastore amenities \
--process extent

go run cmd/railgun/main.go client services add \
--name amenities_centroid \
--title 'Centroid of Amenities' \
--description 'Get the centroid of amenities' \
--datastore amenities \
--process centroid

go run cmd/railgun/main.go client services add \
--name amenities_unique \
--title 'Amenities - Unique Values' \
--description 'Get a list of unique values for a property from a list of amenities' \
--datastore amenities \
--process unique

go run cmd/railgun/main.go client services add \
--name amenities_hist \
--title 'Amenities - Histogram' \
--description 'Get a histogram of values for a property from a list of amenities' \
--datastore amenities \
--process hist

go run cmd/railgun/main.go client services add \
--name amenities_hist_words \
--title 'Amenities - Histogram of Words' \
--description 'For each value of a property, create a histogram of words from a second property from a list of amenities' \
--datastore amenities \
--process hist_words

go run cmd/railgun/main.go client services add \
--name amenities_cuisine_geojson \
--title 'Amenities - Cuisine' \
--description 'Filter amenities by cuisine' \
--datastore amenities \
--process cuisine \
--defaults '{limit: -1}'