#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
set -eu

BASEDIR=~/Downloads/data/geonames

mkdir -p $BASEDIR

#URL='http://download.geonames.org/export/dump/allCountries.zip'

URL='http://download.geonames.org/export/dump/cities1000.zip'

INPUT_HEADER="geonameid,name,asciiname,alternatenames,latitude,longitude,feature_class,feature_code,country_code,cc2,admin1_code,admin2_code,admin3_code,admin4_code,population,elevation,dem,timezone,modification_date"

go run $DIR/../cmd/railgun/main.go process \
--input-uri $URL \
--input-compression zip \
--output-uri "$BASEDIR/cities1000.txt"

go run $DIR/../cmd/railgun/main.go process \
--input-uri "$BASEDIR/cities1000.txt" \
--input-format tsv \
--input-header "$INPUT_HEADER" \
--input-lazy-quotes \
--output-format jsonl \
--stream \
--dfl-uri "$DIR/geonames.dfl" \
--dfl-vars "{dir: '$BASEDIR/', z:6}" \
--output-mkdirs \
--output-format jsonl \
--output-compression 'gzip' \
--output-uri '$dir + "/" + "cities1000.geojsonl.gz"'

exit 0

go run $DIR/../cmd/railgun/main.go process \
--input-uri "$BASEDIR/cities1000.txt" \
--input-format tsv \
--input-header "$INPUT_HEADER" \
--input-lazy-quotes \
--output-format jsonl \
--stream \
--dfl-uri "$DIR/geonames.dfl" \
--dfl-vars "{dir: '$BASEDIR/tiles', z:6}" \
--output-mkdirs \
--output-format jsonl \
--output-compression 'gzip' \
--output-uri '$dir + "/tiles/" + @properties._tile_z + "-" + @properties._tile_x + "-" + @properties._tile_y + ".geojsonl.gz"'