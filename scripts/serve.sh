#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CATALOG_URI=$DIR/../secret/catalog.json

ROOT_PASSWORD=railgun

touch $CATALOG_URI

go run $DIR/../cmd/railgun/main.go serve \
--catalog-uri $CATALOG_URI \
--root-password $ROOT_PASSWORD \
--jwt-public-key-uri $DIR/../secret/keypair.pub \
--jwt-private-key-uri $DIR/../secret/keypair \
--verbose