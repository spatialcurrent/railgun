[![CircleCI](https://circleci.com/gh/spatialcurrent/railgun/tree/master.svg?style=svg)](https://circleci.com/gh/spatialcurrent/railgun/tree/master) [![Go Report Card](https://goreportcard.com/badge/spatialcurrent/railgun)](https://goreportcard.com/report/spatialcurrent/railgun)  [![GoDoc](https://godoc.org/github.com/spatialcurrent/railgun?status.svg)](https://godoc.org/github.com/spatialcurrent/railgun) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/spatialcurrent/railgun/blob/master/LICENSE)

# Railgun

# Description

**Railgun** is a simple and fast data processing tool.  **Railgun** uses:

- [go-reader](https://github.com/spatialcurrent/go-reader) for opening and reading from URIs,
- [go-simple-serializer](https://github.com/spatialcurrent/go-simple-serializer) (GSS) for reading/writing objects to standard formats, and
- [go-dfl](https://github.com/spatialcurrent/go-dfl) for filtering and transforming data.

**Railgun** uses the **Dynamic Filter Language** through **go-dfl**.  See the `*_test` files in the [dfl](https://github.com/spatialcurrent/go-dfl/tree/master/dfl) source folder on GitHub for comprehensive examples of the syntax.

go-reader can read from `stdin`, `http/https`, the local filesystem, [AWS S3](https://aws.amazon.com/s3/), and [HDFS](https://hortonworks.com/apache/hdfs/).

go-simple-serializer (GSS) supports `bson`, `csv`, `tsv`, `hcl`, `hcl2`, `json`, `jsonl`, `properties`, `toml`, `yaml`.  `hcl` and `hcl2` implementation is fragile and very much in `alpha`.

For an interactive demo, see the [railgun notebook](https://beta.observablehq.com/@pjdufour/railgun) on [ObservableHQ](http://observablehq.com). It is very heavy, so only use WiFi.

# Usage

**CLI**

The command line tool, `railgun`, can be used to easily covert data between formats.  We currently support the following platforms.

| GOOS | GOARCH |
| ---- | ------ |
| darwin | amd64 |
| linux | amd64 |
| windows | amd64 |
| linux | arm64 |

Pull requests to support other platforms are welcome!  See the [CLI.md](docs/CLI.md) document for detailed usage and examples.

**Go**

You can install the railgun packages with.


```shell
go get -u -d github.com/spatialcurrent/railguns/...
```

You can then import the main public API with `import "github.com/spatialcurrent/go-simple-serializer/pkg/railgun"` or one of the underlying packages, e.g., `import "github.com/spatialcurrent/go-simple-serializer/pkg/pipeline"`.

See [railgun](https://godoc.org/github.com/spatialcurrent/railgun) in GoDoc for API documentation and examples.

# Releases

**railgun** is currently in **alpha**.  See releases at https://github.com/spatialcurrent/railgun/releases.  See the **Building** section below to build from scratch.

**JavaScript**

- `railgun.global.js`, `railgun.global.js.map` - JavaScript global build  with source map
- `railgun.global.min.js`, `railgun.global.min.js.map` - Minified JavaScript global build with source map
- `railgun.mod.js`, `railgun.mod.js.map` - JavaScript module build  with source map
- `railgun.mod.min.js`, `railgun.mod.min.js.map` - Minified JavaScript module with source map

**Darwin**

- `railgun_darwin_amd64` - CLI for Darwin on amd64 (includes `macOS` and `iOS` platforms)

**Linux**

- `railgun_linux_amd64` - CLI for Linux on amd64
- `railgun_linux_amd64` - CLI for Linux on arm64
- `railgun_linux_amd64.h`, `railgun_linuxamd64.so` - Shared Object for Linux on amd64
- `railgun_linux_armv7.h`, `railgun_linux_armv7.so` - Shared Object for Linux on ARMv7
- `railgun_linux_armv8.h`, `railgun_linux_armv8.so` - Shared Object for Linux on ARMv8

**Windows**

- `railgun_windows_amd64.exe` - CLI for Windows on amd64

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

**Encrypt as Yaml / Decrypt as JSON**

```
# Encrypt secrets.yml and output to secrets.yml.enc
read -s -p 'Password: ' password && echo && railgun_linux_amd64 -input_uri secrets.yml -output_uri secrets.yml.enc -output_passphrase $password
...
# Decrypt secrets.yml.enc and output to stdout
read -s -p 'Password: ' password && echo && railgun_linux_amd64 -input_uri secrets.yml.enc -input_passphrase $password -output_format json
```

# Building

Use `make help` to see help information for each target.

**CLI**

The `make build_cli` script is used to build executables for Linux and Windows.

**JavaScript**

You can compile railgun to pure JavaScript with the `make build_javascript` script.

**Changing Destination**

The default destination for build artifacts is `railgun/bin`, but you can change the destination with an environment variable.  For building on a Chromebook consider saving the artifacts in `/usr/local/go/bin`, e.g., `DEST=/usr/local/go/bin make build_cli`

# Deploying

```
mkdir -p /usr/local/terraform
aws-vault exec default -- terraform init # to download aws provider
cp -R .terraform/plugins/linux_amd64/terraform-provider-aws_v1.43.2_x4 /usr/local/terraform
aws-vault exec default -- terraform init -plugin-dir=/usr/local/terraform
aws-vault exec default -- terraform plan
```

# Testing

**CLI**

To run CLI testes use `make test_cli`, which uses [shUnit2](https://github.com/kward/shunit2).  If you recive a `shunit2:FATAL Please declare TMPDIR with path on partition with exec permission.` error, you can modify the `TMPDIR` environment variable in line or with `export TMPDIR=<YOUR TEMP DIRECTORY HERE>`. For example:

```
TMPDIR="/usr/local/tmp" make test_cli
```

**Go**

To run Go tests use `make test_go` (or `bash scripts/test.sh`), which runs unit tests, `go vet`, `go vet with shadow`, [errcheck](https://github.com/kisielk/errcheck), [ineffassign](https://github.com/gordonklaus/ineffassign), [staticcheck](https://staticcheck.io/), and [misspell](https://github.com/client9/misspell).

**JavaScript**

To run JavaScript tests, first install [Jest](https://jestjs.io/) using `make deps_javascript`, use [Yarn](https://yarnpkg.com/en/), or another method.  Then, build the JavaScript module with `make build_javascript`.  To run tests, use `make test_javascript`.  You can also use the scripts in the `package.json`.

# Contributing

[Spatial Current, Inc.](https://spatialcurrent.io) is currently accepting pull requests for this repository.  We'd love to have your contributions!  Please see [Contributing.md](https://github.com/spatialcurrent/railgun/blob/master/CONTRIBUTING.md) for how to get started.

# License

This work is distributed under the **MIT License**.  See **LICENSE** file.
