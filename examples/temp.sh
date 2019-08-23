#!/bin/bash

#DATASTORE_URI="s3://$RAILGUN_BUCKET/workspace/osm/datastore/amenities/amenities.geojsonl.gz"
#DATASTORE_URI="s3://$RAILGUN_BUCKET/workspace/osm/datastore/amenities/tiles/*.geojsonl.gz"
#DATASTORE_URI="s3://spatialcurrent-data-us-west-2/tiles/osm/pois/8/8-*.geojsonl.gz"
#DATASTORE_URI="~/Downloads/dc_amenities.geojsonl"
#DATASTORE_URI='($z == null) ? "s3://spatialcurrent-data-us-west-2/tiles/osm/pois/8/8-*.geojsonl.gz" : ( ($z < 8) ? null : "s3://spatialcurrent-data-us-west-2/tiles/osm/pois/8/8-" + int64(mul($x, pow(2, sub(8, $z)))) + "-" + int64(mul($y, pow(2, sub(8, $z)))) + ".geojsonl.gz")'

export SERVER=https://railgun.spatialcurrent.io


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

go run cmd/railgun/main.go client services update \
--name chinese_food_geojson \
--title 'Chinese Food' \
--description 'Find Chinese food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{chinese, sichuan, hunan, shanghai}, limit: -1}'
