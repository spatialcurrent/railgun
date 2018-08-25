#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ "$#" -lt 1 ]; then
  echo "Usage VERSION SRC BUCKET"
  exit 1
fi

VERSION=$1
SRC=$2
BUCKET=$3
NAME=railgun

[[ -f "$SRC/$NAME.min.js" ]] && echo "Uploading $SRC/$NAME.min.js" && aws s3 cp $SRC/$NAME.min.js s3://$BUCKET/$NAME/$VERSION/$NAME.min.js
[[ -f "$SRC/$NAME.min.js.map" ]] && echo "Uploading $SRC/$NAME.min.js.map" && aws s3 cp $SRC/$NAME.min.js.map s3://$BUCKET/$NAME/$VERSION/$NAME.min.js.map
[[ -f "$SRC/$NAME.js" ]] && echo "Uploading $SRC/$NAME.js" && aws s3 cp $SRC/$NAME.js s3://$BUCKET/$NAME/$VERSION/$NAME.js
[[ -f "$SRC/$NAME.js.map" ]] && echo "Uploading $SRC/$NAME.js.map" && aws s3 cp $SRC/$NAME.js.map s3://$BUCKET/$NAME/$VERSION/$NAME.js.map
