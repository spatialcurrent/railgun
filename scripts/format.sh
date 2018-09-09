#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
echo "******************"
echo "Formatting $DIR/../railgun"
cd $DIR/../railgun
go fmt
echo "Formatting $DIR/../railgun/cli"
cd $DIR/../railgun/cli
go fmt
echo "Formatting $DIR/../railgun/geo"
cd $DIR/../railgun/geo
go fmt
echo "Formatting $DIR/../railgun/handlers"
cd $DIR/../railgun/handlers
go fmt
echo "Formatting $DIR/../railgun/img"
cd $DIR/../railgun/img
go fmt
echo "Formatting $DIR/../railgun/named"
cd $DIR/../railgun/named
go fmt
echo "Formatting $DIR/../railgun/railgunerrors"
cd $DIR/../railgun/railgunerrors
go fmt
echo "Formatting $DIR/../cmd/railgun"
cd $DIR/../cmd/railgun/
go fmt
echo "Formatting $DIR/../cmd/railgun.js"
cd $DIR/../cmd/railgun.js
go fmt