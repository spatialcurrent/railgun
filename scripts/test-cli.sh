#!/bin/bash

# =================================================================
#
# Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
# Released as open source under the MIT License.  See LICENSE file.
#
# =================================================================

set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export testdata=$(realpath "${DIR}/../testdata")

expectedFormats="bson,csv,fmt,go,gob,hcl,json,jsonl,properties,tags,toml,tsv,yaml"
expectedAlgorithms="bzip2,flate,gzip,none,snappy,zip,zlib"

testFormats() {
  formats=$(railgun formats -f csv)
  assertEquals "unexpected formats" "${expectedFormats}" "${formats}"
}

testAlgorithms() {
  algorithms=$(railgun algorithms -f csv)
  assertEquals "unexpected algorithms" "${expectedAlgorithms}" "${algorithms}"
}

testProcessBatchStdin() {
  local expected='{"hello":"world"}'
  local output=$(echo '{"hello":"world"}' | railgun process --input-uri stdin --input-format json --output-uri stdout --output-format json)
  assertEquals "unexpected output" "${expected}" "${output}"
}

testProcessBatchStdinJSONCSV() {
  local expected='hello\nworld'
  local output=$(echo '{"hello":"world"}' | railgun process --input-uri stdin --input-format json --output-uri stdout --output-format csv)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessStreamJSONLGOBJSONL() {
  local expected='{"hello":"world"}'
  local output=$(railgun process --input-uri "${testdata}/data.json" --input-format jsonl --output-format gob --stream | railgun process --input-format gob --input-type 'map[string]interface{}' --input-trim --output-format jsonl --stream)
  assertEquals "unexpected output" "${expected}" "${output}"
}

testProcessBatchFileJSONCSV() {
  local expected='hello\nworld'
  local output=$(railgun process --input-uri "${testdata}/data.json" --input-format json --output-uri stdout --output-format csv)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessBatchFileCSVJSON() {
  local expected='[{"hello":"world"}]'
  local output=$(railgun process --input-uri "${testdata}/data.csv" --input-format csv --output-uri stdout --output-format json)
  assertEquals "unexpected output" "${expected}" "${output}"
}

testProcessBatchFileJSONCSVGzip() {
  local expected='hello\nworld'
  local output=$(railgun process --input-uri "${testdata}/data.json" --input-format json --output-uri stdout --output-format csv --output-compression gzip | gunzip)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessStreamStdin() {
  local input='{"a":"x"}\n{"b":"y"}\n{"c":"z"}'
  local expected='a=x\nb=y\nc=z'
  local output=$(echo -e "${input}" | railgun process --input-uri stdin --input-format jsonl --output-uri stdout --output-format tags --stream)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessStreamStdinSink() {
  local input='[{"a":"x"},{"b":"y"},{"c":"z"}]'
  local expected='a=x\nb=y\nc=z'
  local output=$(echo -e "${input}" | railgun process --input-uri stdin --input-format json --output-uri stdout --output-format tags --stream)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessStreamStdinMultipleFiles() {
  local input='[{"name":"x"},{"name":"y"},{"name":"z"}]'
  local expected_files='x.tags.gz\ny.tags.gz\nz.tags.gz'
  local output=$(echo -e "${input}" | railgun process --input-uri stdin --input-format json --dfl-vars "{tmpdir: $SHUNIT_TMPDIR}" --output-uri 'format("%s/testresults/%s.tags.gz", $tmpdir, @name)' --output-format go --output-compression gzip --output-mkdirs --stream)
  local actual_files=$(find "${SHUNIT_TMPDIR}/testresults" -type f -printf "%f\n" | sort)
  assertEquals "unexpected output files" "$(echo -e "${expected_files}")" "${actual_files}"
}

testProcessStreamJSONLFmt() {
  local input='{"a":"x"}\n{"b":"y"}\n{"c":"z"}'
  local expected='map[a:x]\nmap[b:y]\nmap[c:z]'
  local output=$(echo -e "${input}" | railgun process --input-format jsonl --output-format fmt --output-format-specifier "%v" --stream)
  assertEquals "unexpected output" "$(echo -e "${expected}")" "${output}"
}

testProcessBatchHCLJSON() {
  local expected='{"data":[{"aws_caller_identity":[{"current":[{}]}]}]}'
  local output=$(echo 'data "aws_caller_identity" "current" {}' | railgun process --input-format hcl --output-format json)
  assertEquals "unexpected output" "${expected}" "${output}"
}

oneTimeSetUp() {
  echo "Using temporary directory at ${SHUNIT_TMPDIR}"
  echo "Reading testdata from ${testdata}"
}

oneTimeTearDown() {
  echo "Tearing Down"
}

# Load shUnit2.
. "${DIR}/shunit2"