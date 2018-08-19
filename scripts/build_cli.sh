#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p $DIR/../bin

echo "******************"
#echo "Formatting $DIR/railgun"
#cd $DIR/../railgun
#go fmt
echo "Formatting $DIR/../cmd/railgun"
cd $DIR/../cmd/railgun
go fmt
echo "Done formatting."
echo "******************"
echo "Building program railgun"
cd $DIR/../bin
####################################################
#echo "Building program for darwin"
#GOTAGS= CGO_ENABLED=1 GOOS=${GOOS} GOARCH=amd64 go build --tags "darwin" -o "railgun_darwin_amd64" github.com/spatialcurrent/railgun/cmd/railgun
#if [[ "$?" != 0 ]] ; then
#    echo "Error building railgun for Darwin"
#    exit 1
#fi
#echo "Executable built at $(realpath $DIR/../bin/railgun_darwin_amd64)"
####################################################
echo "Building program for linux"
GOTAGS= CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build --tags "linux" -o "railgun_linux_amd64" github.com/spatialcurrent/railgun/cmd/railgun
if [[ "$?" != 0 ]] ; then
    echo "Error building railgun for Linux"
    exit 1
fi
echo "Executable built at $(realpath $DIR/../bin/railgun_linux_amd64)"
####################################################
echo "Building program for Windows"
GOTAGS= CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o "railgun_windows_amd64.exe" github.com/spatialcurrent/railgun/cmd/railgun
if [[ "$?" != 0 ]] ; then
    echo "Error building railgun for Windows"
    exit 1
fi
echo "Executable built at $(realpath $DIR/../bin/railgun_windows_amd64.exe)"
