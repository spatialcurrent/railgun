// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/go-simple-serializer/gss"
)

import (
	"github.com/spatialcurrent/railgun/railgun"
)

var GO_RAILGUN_VERSION = "0.0.1"
var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "gzip", "snappy"}
var GO_RAILGUN_FORMATS = []string{"csv", "tsv", "hcl", "hcl2", "json", "jsonl", "properties", "toml", "yaml"}

func printUsage() {
	fmt.Println("Usage: railgun -input_format INPUT_FORMAT -o OUTPUT_FORMAT [-input_uri INPUT_URI] [-input_compression [bzip2|gzip|snappy]] [-h HEADER] [-c COMMENT] [-object_path PATH] [-f FILTER] [-output_path OUTPUT_PATH] [-max MAX_COUNT]")
}

func main() {

	var input_uri string
	var input_compression string
	var input_format string
	var input_header_text string
	var input_comment string

	var object_path string
	var object_filter string

	var output_format string
	var output_path string

	var max_count int

	var version bool
	var help bool

	flag.StringVar(&input_uri, "input_uri", "stdin", "The input uri")
	flag.StringVar(&input_compression, "input_compression", "none", "The input compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	flag.StringVar(&input_format, "input_format", "", "The input format: "+strings.Join(GO_RAILGUN_FORMATS, ", "))
	flag.StringVar(&input_header_text, "h", "", "The input header if the stdin input has no header.")
	flag.StringVar(&input_comment, "c", "", "The input comment character, e.g., #.  Commented lines are not sent to output.")
	flag.StringVar(&object_path, "object_path", "", "The output path")
	flag.StringVar(&object_filter, "f", "", "The output filter")
	flag.StringVar(&output_format, "o", "", "The output format: "+strings.Join(GO_RAILGUN_FORMATS, ", "))
	flag.StringVar(&output_path, "output_path", "", "The output path")
	flag.IntVar(&max_count, "max", -1, "The maximum number of objects to output")
	flag.BoolVar(&version, "version", false, "Prints version to stdout.")
	flag.BoolVar(&help, "help", false, "Print help.")

	flag.Parse()

	if help {
		printUsage()
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Println("Error: Provided no arguments.")
		fmt.Println("Run \"ralgun -help\" for more information.")
		os.Exit(0)
	} else if len(os.Args) == 2 && os.Args[1] == "help" {
		printUsage()
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if version {
		fmt.Println(GO_RAILGUN_VERSION)
		os.Exit(0)
	}

	if len(input_format) == 0 {
		fmt.Println("Error: Provided no -input_format.")
		fmt.Println("Run \"railgun -help\" for more information.")
		os.Exit(1)
	}

	if len(output_format) == 0 {
		fmt.Println("Error: Provided no -output_format.")
		fmt.Println("Run \"railgun -help\" for more information.")
		os.Exit(1)
	}

	var s3_client *s3.S3
	input_reader, err := reader.OpenResource(input_uri, input_compression, false, s3_client)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error opening resource from uri "+input_uri))
	}
	input_bytes, err := input_reader.ReadAll()
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error reading from resource"))
	}
	input_string := string(input_bytes)

	input_header := make([]string, 0)
	if len(input_header_text) > 0 {
		input_header = strings.Split(input_header_text, ",")
	}

	var root dfl.Node
	var funcs dfl.FunctionMap
	if len(object_filter) > 0 {
		n, err := dfl.Parse(object_filter)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error parsing filter expression."))
		}
		root = n.Compile()
		funcs = dfl.NewFuntionMapWithDefaults()
	}

	input_object := gss.NewObject(input_string, input_format)
	err = gss.Deserialize(input_string, input_format, input_header, input_comment, &input_object)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error deserializing input using format "+input_format))
	}

	output, err := railgun.Process(input_object, object_path, root, funcs, max_count, output_path)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error processing"))
	}

	output_string, err := gss.Serialize(output, output_format)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error converting"))
	}
	fmt.Println(output_string)
}
