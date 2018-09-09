#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
DEST=${1:-$DIR/../bin}

mkdir -p $DEST

echo "******************"
echo "Removing cache"
rm -fr  ~/go/pkg/*_js
echo "******************"
echo "Building Javascript for Railgun"
cd $DEST
gopherjs build -o railgun.js github.com/spatialcurrent/railgun/cmd/railgun.js
if [[ "$?" != 0 ]] ; then
    echo "Error building Javascript artificats for Railgun"
    exit 1
fi
gopherjs build -m -o railgun.min.js github.com/spatialcurrent/railgun/cmd/railgun.js
if [[ "$?" != 0 ]] ; then
    echo "Error building Javascript artificats for Railgun"
    exit 1
fi
echo "JavaScript artifacts built at $DEST"
