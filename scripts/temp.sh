#!/bin/bash

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

go run cmd/railgun/main.go client processes add \
--name medical \
--title Medical Services \
--description 'Filter a list of GeoJSON features for medical services' \
--tags '[geojson]' \
--expression 'filter(@, "(@properties?.amenity != null) and (@properties.amenity in [clinic,doctors,hospital])") | (($limit > 0) ? limit(@, $limit) : @) | {type:FeatureCollection, features:@, numberOfFeatures: len(@)}'

go run cmd/railgun/main.go client services add \
--name medical_services_geojson \
--title 'Medical Services' \
--description 'Find medical services' \
--tags '[medical, healthcare, geojson]' \
--datastore amenities \
--process medical \
--defaults '{limit: -1}'
