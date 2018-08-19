[![Build Status](https://travis-ci.org/spatialcurrent/railgun.svg)](https://travis-ci.org/spatialcurrent/railgun) [![GoDoc](https://godoc.org/github.com/spatialcurrent/railgun?status.svg)](https://godoc.org/github.com/spatialcurrent/railgun)

# Railgun

# Description

**Railgun** is a simple and fast data processing tool.  **Railgun** uses [go-simple-serializer](https://github.com/spatialcurrent/go-simple-serializer) (GSS) for reading/writing objects to standard formats.  **Railgun** uses [go-dfl](https://github.com/spatialcurrent/go-dfl) for filtering.

GSS supports `bson`, `csv`, `tsv`, `hcl`, `hcl2`, `json`, `jsonl`, `properties`, `toml`, `yaml`.  `hcl` and `hcl2` implementation is fragile and very much in `alpha`.

# Usage

**CLI**

You can use the command line tool to convert between formats.

```
Usage: railgun -input_format INPUT_FORMAT -o OUTPUT_FORMAT [-input_uri INPUT_URI] [-input_compression [bzip2|gzip|snappy]] [-h HEADER] [-c COMMENT] [-object_path PATH] [-f FILTER] [-output_path OUTPUT_PATH] [-max MAX_COUNT]
Options:
  -c string
    	The input comment character, e.g., #.  Commented lines are not sent to output.
  -f string
    	The output filter
  -h string
    	The input header if the stdin input has no header.
  -help
    	Print help.
  -input_compression string
    	The input compression: none, gzip, snappy (default "none")
  -input_format string
    	The input format: csv, tsv, hcl, hcl2, json, jsonl, properties, toml, yaml
  -input_uri string
    	The input uri (default "stdin")
  -max int
    	The maximum number of objects to output (default -1)
  -o string
    	The output format: csv, tsv, hcl, hcl2, json, jsonl, properties, toml, yaml
  -object_path string
    	The output path
  -output_path string
    	The output path
  -version
    	Prints version to stdout.
```

**Go**

You can import **railgun** as a library with:

```go
import (
  "github.com/spatialcurrent/go-railgun/railgun"
)
```

The `Process` function is the core functions to use.

```go
...
  output_object, err := railgun.Process(input_object, object_path, filter, funcs, max_count, output_path)
...
  output_string, err := gss.Serialize(output_object, output_format)
...
```

# Releases

**Railgun** is currently in **alpha**.  See releases at https://github.com/spatialcurrent/railgun/releases.

# Examples

TBD

# Building

**CLI**

The `build_cli.sh` script is used to build executables for Linux and Windows.

# Contributing

[Spatial Current, Inc.](https://spatialcurrent.io) is currently accepting pull requests for this repository.  We'd love to have your contributions!  Please see [Contributing.md](https://github.com/spatialcurrent/railgun/blob/master/CONTRIBUTING.md) for how to get started.

# License

This work is distributed under the **MIT License**.  See **LICENSE** file.
