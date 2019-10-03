# CLI

- [Usage](#usage) - basic usage
- [Algorithms](#algorithms) - list of supported compression algorithms
- [Formats](#formats) - list of supported file formats
- [Platforms](#platforms) - list of supported platforms
- [Releases](#releases) - where to find an executable
- [Examples](#examples)  - detailed usage exampels
- [Building](#building) - how to build the CLI
- [Testing](#testing) - test the CLI
- [Troubleshooting](#Troubleshooting) - how to troubleshoot common errors

## Usage

The command line tool, `railgun`, can be used to easily covert data between file formats and compression algorithms.

To launch the http serve use the `serve` command.

```shell
railgun serve [flags]
```

To process data use the `process` command.


```shell
railgun process [flags]
```

For more information use the help flag.

```shell
railgun --help
```

## Algorithms

The following compression algorithms are supported.  Pull requests to support other compression algorithms are welcome!

| Algorithm | Read |  Write | Stream | Description |
| ---- | ------ | ------ | ------ | ------ |
| bzip2 | ✓ | - | ✓ | [bzip2](https://en.wikipedia.org/wiki/Bzip2) |
| flate | ✓ | ✓ | ✓ | [DEFLATE Compressed Data Format](https://tools.ietf.org/html/rfc1951) |
| gzip | ✓ | ✓ | ✓ | [gzip](https://en.wikipedia.org/wiki/Gzip) |
| snappy | ✓ | ✓ | ✓ | [snappy](https://github.com/google/snappy) |
| zip | ✓ | - | - | [zip](https://en.wikipedia.org/wiki/Zip_%28file_format%29) |
| zlib | ✓ | ✓ | ✓ | [zlib](https://en.wikipedia.org/wiki/Zlib) |


## Formats

The following file formats are supported.  Pull requests to support other formats are welcome!

| Format | Read |  Write | Stream | Description |
| ---- | ------ |  ------ | ------ | ------ |
| bson | ✓ | ✓ | - | [Binary JSON](https://en.wikipedia.org/wiki/BSON) |
| csv | ✓ | ✓ | ✓ | [Comma-Separated Values](https://en.wikipedia.org/wiki/Comma-separated_values) |
| fmt | - | ✓ | ✓ | [fmt](https://godoc.org/fmt) |
| go | - | ✓ | ✓ | Go (format specifier: "%#v") |
| gob | ✓ | ✓ | ✓ | [gob](https://godoc.org/encoding/gob) |
| hcl | ✓ | - | - | [HashiCorp Configuration Language](https://github.com/hashicorp/hcl) |
| json | ✓ | ✓ | - | [JSON](http://json.org/) |
| jsonl | ✓ | ✓ | ✓ | [JSON Lines](http://jsonlines.org/) |
| properties | ✓ | ✓ | - |[Properties](https://en.wikipedia.org/wiki/.properties) |
| tags | ✓ | ✓ | ✓ | single-line series of key=value tags |
| toml | ✓ | ✓ | - | [TOML](https://github.com/toml-lang/toml) |
| tsv | ✓ | ✓ | ✓ |[ Tab-Separated Values](https://en.wikipedia.org/wiki/Tab-separated_values) |
| yaml | ✓ | ✓ | - | [YAML](https://yaml.org/) |


## Platforms

The following platforms are supported.  Pull requests to support other platforms are welcome!

| GOOS | GOARCH |
| ---- | ------ |
| darwin | amd64 |
| linux | amd64 |
| windows | amd64 |
| linux | arm64 |

## Releases

**railgun** is currently in **alpha**.  See releases at [https://github.com/spatialcurrent/railgun/releases]([https://github.com/spatialcurrent/railgun/releases].  See the **Building** section below to build from scratch.

**Darwin**

- `railgun_darwin_amd64` - CLI for Darwin on amd64 (includes `macOS` and `iOS` platforms)

**Linux**

- `railgun_linux_amd64` - CLI for Linux on amd64
- `railgun_linux_amd64` - CLI for Linux on arm64

**Windows**

- `railgun_windows_amd64.exe` - CLI for Windows on amd64

## Examples

**Search for Cuisine**

```
~/go/src/github.com/spatialcurrent/go-osm/bin/osm_linux_amd64 -input_uri 'http://download.geofabrik.de/north-america/us/district-of-columbia-latest.osm.bz2' -ways_to_nodes -output_format geojsonl -filter_keys_keep amenity -output_uri stdout | railgun -input_format jsonl  -output_format json -dfl_file ~/go/src/github.com/spatialcurrent/railgun/examples/mexican.dfl -output_uri mexican.json
```

**Encrypt as Yaml / Decrypt as JSON**

```
# Encrypt secrets.yml and output to secrets.yml.enc
read -s -p 'Password: ' password && echo && railgun_linux_amd64 -input_uri secrets.yml -output_uri secrets.yml.enc -output_passphrase $password
...
# Decrypt secrets.yml.enc and output to stdout
read -s -p 'Password: ' password && echo && railgun_linux_amd64 -input_uri secrets.yml.enc -input_passphrase $password -output_format json
```

## Building

Use `make build_cli` to build executables for Linux and Windows.

**Changing Destination**

The default destination for build artifacts is `railgun/bin`, but you can change the destination with an environment variable.  For building on a Chromebook consider saving the artifacts in `/usr/local/go/bin`, e.g., `DEST=/usr/local/go/bin make build_cli`

## Testing

To run CLI testes use `make test_cli`, which uses [shUnit2](https://github.com/kward/shunit2).  If you recive a `shunit2:FATAL Please declare TMPDIR with path on partition with exec permission.` error, you can modify the `TMPDIR` environment variable in line or with `export TMPDIR=<YOUR TEMP DIRECTORY HERE>`. For example:

```
TMPDIR="/usr/local/tmp" make test_cli
```

## Troubleshooting

### no such file or directory

#### Example

```text
Error: error processing as stream: error writing buffers to files: error opening output file for path %q: error opening file for writing at path %q: open %q: no such file or directory
```

#### Solution

This error typically occurs when a parent directory of an output file does not exist.  Use the `--output-mkdirs` command line flag to allow railgun to create parent directories for output files as needed.

