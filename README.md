[![Build Status](https://travis-ci.org/spatialcurrent/railgun.svg)](https://travis-ci.org/spatialcurrent/railgun) [![Go Report Card](https://goreportcard.com/badge/spatialcurrent/railgun)](https://goreportcard.com/report/spatialcurrent/railgun)  [![GoDoc](https://godoc.org/github.com/spatialcurrent/railgun?status.svg)](https://godoc.org/github.com/spatialcurrent/railgun) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/spatialcurrent/railgun/blob/master/LICENSE)

# Railgun

# Description

**Railgun** is a simple and fast data processing tool.  **Railgun** uses:
- [go-reader](https://github.com/spatialcurrent/go-reader) for opening and reading from URIs,
- [go-simple-serializer](https://github.com/spatialcurrent/go-simple-serializer) (GSS) for reading/writing objects to standard formats, and
- [go-dfl](https://github.com/spatialcurrent/go-dfl) for filtering and transforming data.

go-reader can read from `stdin`, `http/https`, the local filesystem, [AWS S3](https://aws.amazon.com/s3/), and [HDFS](https://hortonworks.com/apache/hdfs/).

go-simple-serializer (GSS) supports `bson`, `csv`, `tsv`, `hcl`, `hcl2`, `json`, `jsonl`, `properties`, `toml`, `yaml`.  `hcl` and `hcl2` implementation is fragile and very much in `alpha`.

# Usage

**CLI**

You can use the command line tool to process data.

```
Usage: railgun -input_format INPUT_FORMAT -o OUTPUT_FORMAT [-input_uri INPUT_URI] [-input_compression [bzip2|gzip|snappy]] [-h HEADER] [-c COMMENT] [-object_path PATH] [-dfl_exp DFL_EXPRESSION] [-dfl_file DFL_FILE] [-output_path OUTPUT_PATH] [-max MAX_COUNT]
Options:
  -aws_access_key_id string
    	Defaults to value of environment variable AWS_ACCESS_KEY_ID
  -aws_default_region string
    	Defaults to value of environment variable AWS_DEFAULT_REGION.
  -aws_secret_access_key string
    	Defaults to value of environment variable AWS_SECRET_ACCESS_KEY.
  -aws_session_token string
    	Defaults to value of environment variable AWS_SESSION_TOKEN.
  -c string
    	The input comment character, e.g., #.  Commented lines are not sent to output.
  -dfl_exp string
    	Process using dfl expression
  -dfl_file string
    	Process using dfl file.
  -h string
    	The input header if the stdin input has no header.
  -hdfs_name_node string
    	Defaults to value of environment variable HDFS_DEFAULT_NAME_NODE.
  -help
    	Print help.
  -input_compression string
    	The input compression: none, bzip2, gzip, snappy (default "none")
  -input_format string
    	The input format: bson, csv, tsv, hcl, hcl2, json, jsonl, properties, toml, yaml
  -input_reader_buffer_size int
    	The input reader buffer size (default 4096)
  -input_uri string
    	The input uri (default "stdin")
  -max int
    	The maximum number of objects to output (default -1)
  -output_format string
    	The output format: bson, csv, tsv, hcl, hcl2, json, jsonl, properties, toml, yaml
  -output_uri string
    	The output uri (default "stdout")
  -version
    	Prints version to stdout.
```

# Releases

**Railgun** is currently in **alpha**.  See releases at https://github.com/spatialcurrent/railgun/releases.

# Examples

**Search for Cuisine**

```
~/go/src/github.com/spatialcurrent/go-osm/bin/osm_linux_amd64 -input_uri 'http://download.geofabrik.de/north-america/us/district-of-columbia-latest.osm.bz2' -ways_to_nodes -output_format geojsonl -filter_keys_keep amenity -output_uri stdout | railgun -input_format jsonl  -output_format json -dfl_file ~/go/src/github.com/spatialcurrent/railgun/examples/mexican.dfl -output_uri mexican.json
```

**Tsunami Feed**

```
const pipeline = ["filter(@features, '(@properties?.tsunami != null) and (@properties.tsunami == 1)')", "sort(@, '@properties?.mag', true)", "map(@, '@properties?.place ?: \"\"')", "limit(@, 10)"];
(await fetch("https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/2.5_month.geojson")).json().then(earthquakes => {
  result = railgun.process(earthquakes, {"dfl": pipeline, "output_format": "yaml"});
  console.log(result);
})
```

# Building

**CLI**

The `build_cli.sh` script is used to build executables for Linux and Windows.

**JavaScript**

You can compile GSS to pure JavaScript with the `scripts/build_javascript.sh` script.

**Changing Destination**

The default destination for build artifacts is `railgun/bin`, but you can change the destination with a CLI argument.  For building on a Chromebook consider saving the artifacts in `/usr/local/go/bin`, e.g., `bash scripts/build_cli.sh /usr/local/go/bin`

# Contributing

[Spatial Current, Inc.](https://spatialcurrent.io) is currently accepting pull requests for this repository.  We'd love to have your contributions!  Please see [Contributing.md](https://github.com/spatialcurrent/railgun/blob/master/CONTRIBUTING.md) for how to get started.

# License

This work is distributed under the **MIT License**.  See **LICENSE** file.
