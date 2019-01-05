#!/bin/bash

export SERVER=https://railgun.spatialcurrent.io

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

export JWT_TOKEN=$(echo $RAILGUN_AUTH_RESP | jq -r .token)

echo "****************************"
echo "Using JWT Token:"
echo $JWT_TOKEN
echo "****************************"

go run cmd/railgun/main.go client services update \
--name breweries_geojson \
--title 'Breweries' \
--description 'Find breweries' \
--tags '[beer, geojson]' \
--datastore amenities \
--process crafts \
--defaults '{crafts:{brewery}, limit: -1}'

go run cmd/railgun/main.go client services update \
--name ethiopian_food_geojson \
--title 'Ethiopian Food' \
--description 'Find Ethiopian food' \
--tags '[cuisine, geojson]' \
--datastore amenities \
--process cuisines \
--defaults '{cuisines:{ethiopian}, limit: -1}'

go run cmd/railgun/main.go client services update \
--name distilleries_geojson \
--title 'Distilleries' \
--description 'Find distilleries' \
--tags '[craft, liquor, geojson]' \
--datastore amenities \
--process crafts \
--defaults '{crafts:{distillery}, limit: -1}'