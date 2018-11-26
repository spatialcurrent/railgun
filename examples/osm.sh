#!/bin/bash

#DATASTORE_URI="s3://$RAILGUN_BUCKET/workspace/osm/datastore/amenities/amenities.geojsonl.gz"
#DATASTORE_URI="s3://$RAILGUN_BUCKET/workspace/osm/datastore/amenities/tiles/*.geojsonl.gz"
DATASTORE_URI="s3://$RAILGUN_BUCKET/tiles/osm/pois/8/8-*.geojsonl.gz"
#DATASTORE_URI="~/Downloads/dc_amenities.geojsonl"



#if [[ -z "${RAILGUN_BUCKET}" ]]; then
#  echo "Requires RAILGUN_BUCKET environment variable to be defined"
#  exit 1
#fi

echo -n 'Username: '
read username
echo
echo -n 'Password: '
read -s password
echo
echo "****************************"
echo "* authenticating with user $username ..."

RAILGUN_AUTH_RESP=$(USERNAME=$username PASSWORD=$password go run cmd/railgun/main.go client authenticate --error-destination stdout)
RAILGUN_AUTH_SUCCESS=$(echo $RAILGUN_AUTH_RESP | jq -r .success)
if [[ "$RAILGUN_AUTH_SUCCESS" != "true" ]]; then
  echo "Authentication failed."
  exit 1
fi
echo "* authentication successful"

# don't set pipefail too early, otherwise we can't inspect response from client authenticate
#set -eo pipefail

# sets JWT_TOKEN for use by following commands
export JWT_TOKEN=$(echo $RAILGUN_AUTH_RESP | jq -r .token)

echo "****************************"
echo "Using JWT Token:"
echo $JWT_TOKEN
echo "****************************"

go run cmd/railgun/main.go client workspaces add \
--name osm \
--title osm \
--description 'Workspace for OpenStreetMap data'

go run cmd/railgun/main.go client datastores add \
--workspace osm \
--name amenities \
--title Amenities \
--description 'Amenities from OpenStreetMap' \
--extent '[-77.5195609,38.8099849,-76.9102596,39.1546259]' \
--uri $DATASTORE_URI

go run cmd/railgun/main.go client layers add \
--name amenities \
--title Amenities \
--description 'OpenStreetMap amenities' \
--tags '[amenities]' \
--datastore amenities \
--expression ''

go run cmd/railgun/main.go client layers add \
--name medical \
--title Medical Services \
--description 'Filter a list of GeoJSON features for medical services' \
--tags '[amenities, medical, healthcare]' \
--datastore amenities \
--expression '(@properties?.amenity != null) and (@properties.amenity in [clinic,doctors,hospital])'

go run cmd/railgun/main.go client layers add \
--name japanese_food \
--title 'Japanese Food' \
--description 'Find Japanese food' \
--tags '[amenities, food, cuisine, japanese]' \
--datastore amenities \
--expression '((@properties?.cuisine != null) and (@properties?.cuisine iin $cuisines)) or ((@properties?.name != null) and (intersects($cuisines, set(split(lower(@properties.name),` `)))))' \
--defaults '{cuisines:{sushi, japanese}, limit: -1}'

go run cmd/railgun/main.go client layers add \
--name thai_food \
--title 'Thai Food' \
--description 'Find Thai food' \
--tags '[amenities, food, cuisine, thai]' \
--datastore amenities \
--expression '(@properties?.cuisine ilike $cuisine) or ((@properties?.name != null) and ($cuisine iin split(@properties.name,` `)))' \
--defaults '{cuisine:thai, limit: -1}'

#go run cmd/railgun/main.go client datastores add \
#--workspace osm \
#--name amenities_tiles \
#--title Amenities - Tiles \
#--description 'Amenities from OpenStreetMap as Tiles' \
#--format 'jsonl' \
#--compression 'snappy' \
#--uri '($tileZ := 10) | (@z <  $tileZ) ? null : "~/Downloads/data/dc_amenities/10-" + int64(mul(@x, pow(2, sub($tileZ, @z)))) + "-" + int64(mul(@y, pow(2, sub($tileZ, @z)))) + ".geojsonl.sz"' \
#--extent '[-77.5195609,38.8099849,-76.9102596,39.1546259]'

#go run cmd/railgun/main.go client layers add \
#--name amenities \
#--title Amenities \
#--description 'Amenities from OpenStreetMap' \
#--datastore amenities_tiles

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
--description 'Filter a list of GeoJSON features by cuisines' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.cuisine ilike $cuisine) or ((@properties?.name != null) and ($cuisine iin split(@properties.name,` `)))", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name medical \
--title Medical Services \
--description 'Filter a list of GeoJSON features for medical services' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity in [clinic,doctors,hospital])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name crafts \
--title Crafts \
--description 'Filter a list of GeoJSON features by crafts.' \
--tags '[geojson]' \
--expression 'filter(@, "((@properties?.craft != null) and (@properties?.craft iin $crafts)) or ((@properties?.name != null) and (intersects($crafts , set(split(lower(@properties.name),` `)))))", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name cuisines \
--title Cuisines \
--description 'Filter a list of GeoJSON features by cuisines' \
--tags '[geojson]' \
--expression 'filter(@, "((@properties?.cuisine != null) and (@properties?.cuisine iin $cuisines)) or ((@properties?.amenity != embassy) and (@properties?.name != null) and (intersects($cuisines , set(split(lower(@properties.name),` `)))))", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name gas_stations \
--title Gas Stations \
--description 'Filter a list of GeoJSON features for gas stations' \
--tags '[geojson]' \
--expression '($c := ($c ?: filter(@, "(@properties?.amenity != null) and (@properties.amenity in [fuel])"))) | $c | (($limit > 0) ? limit(@, $limit) : @) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name charging_stations \
--title Charging Stations \
--description 'Filter a list of GeoJSON features for charging stations for electric vehicles.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity in [charging_station])") | (($limit > 0) ? limit(@, $limit) : @) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name lodging \
--title Lodging \
--description 'Filter a list of GeoJSON features for lodging, e.g., hotels, motels, etc.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.tourism != null) and (@properties.tourism iin [hotel, motel])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name airports \
--title Airports \
--description 'Filter a list of GeoJSON features for airports.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.aeroway != null) and (@properties.aeroway iin [aerodrome])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name helipads \
--title Helipads \
--description 'Filter a list of GeoJSON features for helipads.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.aeroway != null) and (@properties.aeroway iin [helipad, heliport])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name ice_skating \
--title Ice Skating \
--description 'Filter a list of GeoJSON features for locations for ice skating.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.leisure != null) and (@properties.leisure iin [ice_rink])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name workout \
--title Workout \
--description 'Filter a list of GeoJSON features for locations for working out.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity iin [fitness_center, gym])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name cafes \
--title Cafes \
--description 'Filter a list of GeoJSON features for cafes.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity iin [cafe])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name post_offices \
--title Post Offices \
--description 'Filter a list of GeoJSON features for post offices.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity in [post_office])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client processes add \
--name parks \
--title Parks \
--description 'Filter a list of GeoJSON features for parks.' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.leisure != null) and (@properties.leisure iin [park])", $limit) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

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
--description 'For each value of a property, create a histogram of words from a second property from a list of amenities' --datastore amenities \
--process hist_words

go run cmd/railgun/main.go client jobs add \
--name amenities_hist_cuisine \
--title amenities_hist_cuisine \
--description amenities_hist_cuisine \
--service amenities_hist \
--variables '{name: cuisine}'

go run cmd/railgun/main.go client jobs add \
--name amenities_unique_cuisine \
--title amenities_unique_cuisine \
--description amenities_unique_cuisine \
--service amenities_unique \
--variables '{name: cuisine}'

go run cmd/railgun/main.go client jobs add \
--name amenities_centroid_cuisine \
--title amenities_centroid_cuisine \
--description amenities_centroid_cuisine \
--service amenities_centroid \
--variables '{name: cuisine}'

go run cmd/railgun/main.go client workflows add \
--name osm \
--title OSM \
--description OpenStretMap Workflow \
--jobs '[amenities_unique_cuisine, amenities_centroid_cuisine, amenities_hist_cuisine]'

go run cmd/railgun/main.go client services add \
--name thai_food_geojson \
--title 'Thai Food' \
--description 'Find Thai food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisine \
--defaults '{cuisine:thai, limit: -1}'

go run cmd/railgun/main.go client services add \
--name japanese_food_geojson \
--title 'Japanese Food' \
--description 'Find Japanese food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{sushi, japanese}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name medical_services_geojson \
--title 'Medical Services' \
--description 'Find medical services' \
--tags '[medical, healthcare, geojson]' \
--datastore amenities \
--process medical \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name gas_stations_geojson \
--title 'Gas Stations' \
--description 'Find gas stations' \
--tags '[transit, transportation, geojson]' \
--datastore amenities \
--process gas_stations \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name charging_stations_geojson \
--title 'Charging Stations' \
--description 'Find charging stations for electric vehicles.' \
--tags '[transit, transportation, geojson]' \
--datastore amenities \
--process charging_stations \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name lodging_geojson \
--title 'Lodging' \
--description 'Find lodging, etc., hotels, motels, etc.' \
--tags '[lodging, hotels, motels, geojson]' \
--datastore amenities \
--process lodging \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name airports_geojson \
--title 'Airports' \
--description 'Find airports.' \
--tags '[transit, transportation, geojson]' \
--datastore amenities \
--process airports \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name helipads_geojson \
--title 'Helipads' \
--description 'Find helipads for use by a helicopter.' \
--tags '[transit, transportation, geojson]' \
--datastore amenities \
--process helipads \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name ice_skating_geojson \
--title 'Ice Skating' \
--description 'Find locations for ice skating.' \
--tags '[leisure, geojson]' \
--datastore amenities \
--process ice_skating \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name workout_geojson \
--title 'Workout' \
--description 'Find locations for working out.' \
--tags '[fitness, geojson]' \
--datastore amenities \
--process workout \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name cafes_geojson \
--title 'Cafes' \
--description 'Find cafes.' \
--tags '[coffee, geojson]' \
--datastore amenities \
--process cafes \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name mexican_food_geojson \
--title 'Mexican Food' \
--description 'Find Mexican food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{burrito, taco, mexican}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name chinese_food_geojson \
--title 'Chinese Food' \
--description 'Find Chinese food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{chinese}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name ethiopian_food_geojson \
--title 'Ethiopian Food' \
--description 'Find Ethiopian food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{ethiopian}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name vietnamese_food_geojson \
--title 'Vietnamese Food' \
--description 'Find Vietnamese food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{vietnamese}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name breweries_geojson \
--title 'Breweries' \
--description 'Find breweries' \
--tags '[beer, geojson]' \
--datastore amenities \
--process crafts \
--defaults '{crafts:{brewery}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name distilleries_geojson \
--title 'Distilleries' \
--description 'Find distilleries' \
--tags '[craft, liquor, geojson]' \
--datastore amenities \
--process distilleries \
--defaults '{crafts:{distillery}, limit: -1}'

go run cmd/railgun/main.go client services add \
--name post_offices_geojson \
--title 'Post Offices' \
--description 'Find post offices.' \
--tags '[mail, geojson]' \
--datastore amenities \
--process post_offices \
--defaults '{limit: -1}'

go run cmd/railgun/main.go client services add \
--name parks_geojson \
--title 'Parks' \
--description 'Find parks.' \
--tags '[recreation, leisure, geojson]' \
--datastore amenities \
--process parks \
--defaults '{limit: -1}'