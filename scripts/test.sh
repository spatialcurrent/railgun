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
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun
echo "Testing $DIR/../railgun/cli"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/ci
echo "Testing $DIR/../railgun/geo"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/geo
echo "Testing $DIR/../railgun/handlers"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/handlers
echo "Testing $DIR/../railgun/img"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/img
echo "Testing $DIR/../railgun/named"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/named
echo "Testing $DIR/../railgun/railgunerrors"
gometalinter --misspell-locale=US --disable-all --enable=misspell --enable=vet --enable=ineffassign --enable=gosec $DIR/../railgun/railgunerrors