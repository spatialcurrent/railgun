#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

LDFLAGS="-X main.gitBranch=$(git branch | grep \* | cut -d ' ' -f2) -X main.gitCommit=$(git rev-list -1 HEAD)"

CATALOG_URI=$DIR/../secret/catalog.json

ROOT_PASSWORD=railgun

touch $CATALOG_URI

export RUNTIME_MAX_PROCS=0
export HTTP_ADDRESS=0.0.0.0:8080
export HTTP_MIDDLEWARE_DEBUG=1
export HTTP_MIDDLEWARE_RECOVER=1
export HTTP_MIDDLEWARE_GZIP=1
export HTTP_MIDDLEWARE_CORS=0
export HTTP_GRACEFUL_SHUTDOWN=1
export VERBOSE=1

go run -ldflags "$LDFLAGS" $DIR/../cmd/railgun/main.go serve \
--catalog-uri $CATALOG_URI \
--root-password $ROOT_PASSWORD \
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