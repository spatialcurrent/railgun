#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
set -eu
echo "******************"
echo "Running unit tests"
#cd $DIR/../railgun
#go test
echo "******************"
echo "Using gometalinter with misspell, vet, ineffassign, and gosec"
echo "Testing $DIR/../railgun"
# removed --enable=misspell, until I found out how to give custom dictionary
gometalinter --deadline=60s --misspell-locale=US --disable-all  --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun
echo "Testing $DIR/../railgun/catalog"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/catalog
echo "Testing $DIR/../railgun/cli"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/cli
echo "Testing $DIR/../railgun/geo"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/geo
echo "Testing $DIR/../railgun/handlers"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/handlers
echo "Testing $DIR/../railgun/logger"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/logger
echo "Testing $DIR/../railgun/middleware"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/middleware
echo "Testing $DIR/../railgun/img"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/img
echo "Testing $DIR/../railgun/named"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/named
echo "Testing $DIR/../railgun/railgunerrors"
gometalinter --deadline=60s --misspell-locale=US --disable-all --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/railgunerrors