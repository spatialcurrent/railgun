#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [[ ! -f $DIR/../secret/keypair.pub ]]; then
  echo "You are missing a keypair for JWT.  Please run gen_keypair.sh."
  exit 1
fi

LDFLAGS="-X main.gitBranch=$(git branch | grep \* | cut -d ' ' -f2) -X main.gitCommit=$(git rev-list -1 HEAD)"

#export CATALOG_URI=$DIR/../secret/catalog.remote.json
#export CATALOG_URI=$DIR/../secret/prod.json
export CATALOG_URI=$DIR/../secret/local.json
export ROOT_PASSWORD=railgun
export RUNTIME_MAX_PROCS=0
export HTTP_ADDRESS=0.0.0.0:8080
export HTTP_MIDDLEWARE_DEBUG=1
export HTTP_MIDDLEWARE_RECOVER=1
export HTTP_MIDDLEWARE_GZIP=1
export HTTP_MIDDLEWARE_CORS=1
export HTTP_GRACEFUL_SHUTDOWN=1
export VERBOSE=1

mkdir -p $DIR/../secret
if [[ ! -f $CATALOG_URI ]]; then
  echo "Catalog Missing.  Creating new catalog"
  touch $CATALOG_URI
fi

go run -ldflags "$LDFLAGS" $DIR/../cmd/railgun/main.go serve \
--jwt-public-key-uri $DIR/../secret/keypair.pub \
--jwt-private-key-uri $DIR/../secret/keypair \
--http-timeout-read 60s \
--http-timeout-write 60s

#go run $DIR/../cmd/railgun/main.go serve \
#--catalog-uri $CATALOG_URI \
#--root-password $ROOT_PASSWORD \
#--jwt-public-key-uri $DIR/../secret/keypair.pub \
#--jwt-private-key-uri $DIR/../secret/keypair \
#--http-timeout-read 60s \
#--http-timeout-write 60s
