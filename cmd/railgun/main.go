// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

import (
	"github.com/colinmarc/hdfs"
	"github.com/pkg/errors"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/railgun/railgun"
)

var GO_RAILGUN_COMPRESSION_ALGORITHMS = []string{"none", "bzip2", "gzip", "snappy"}
var GO_RAILGUN_FORMATS = []string{"bson", "csv", "tsv", "hcl", "hcl2", "json", "jsonl", "properties", "toml", "yaml"}

func printUsage() {
	fmt.Println("Usage: railgun -input_format INPUT_FORMAT -o OUTPUT_FORMAT [-input_uri INPUT_URI] [-input_compression [bzip2|gzip|snappy]] [-h HEADER] [-c COMMENT] [-object_path PATH] [-dfl_exp DFL_EXPRESSION] [-dfl_file DFL_FILE] [-output_path OUTPUT_PATH] [-max MAX_COUNT]")
}

func connect_to_aws(aws_access_key_id string, aws_secret_access_key string, aws_session_token string, aws_region string) *session.Session {
	aws_session := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, aws_session_token),
			MaxRetries:  aws.Int(3),
			Region:      aws.String(aws_region),
		},
	}))
	return aws_session
}

func main() {

	var aws_default_region string
	var aws_access_key_id string
	var aws_secret_access_key string
	var aws_session_token string
	var hdfs_name_node string

	var input_uri string
	var input_compression string
	var input_format string
	var input_header_text string
	var input_comment string
	var input_reader_buffer_size int

	var dfl_exp string
	var dfl_file string

	var output_uri string
	var output_format string

	var max_count int

	var version bool
	var verbose bool
	var help bool

	flag.StringVar(&aws_default_region, "aws_default_region", "", "Defaults to value of environment variable AWS_DEFAULT_REGION.")
	flag.StringVar(&aws_access_key_id, "aws_access_key_id", "", "Defaults to value of environment variable AWS_ACCESS_KEY_ID")
	flag.StringVar(&aws_secret_access_key, "aws_secret_access_key", "", "Defaults to value of environment variable AWS_SECRET_ACCESS_KEY.")
	flag.StringVar(&aws_session_token, "aws_session_token", "", "Defaults to value of environment variable AWS_SESSION_TOKEN.")

	flag.StringVar(&hdfs_name_node, "hdfs_name_node", "", "Defaults to value of environment variable HDFS_DEFAULT_NAME_NODE.")

	flag.StringVar(&input_uri, "input_uri", "stdin", "The input uri")
	flag.StringVar(&input_compression, "input_compression", "none", "The input compression: "+strings.Join(GO_RAILGUN_COMPRESSION_ALGORITHMS, ", "))
	flag.StringVar(&input_format, "input_format", "", "The input format: "+strings.Join(GO_RAILGUN_FORMATS, ", "))
	flag.StringVar(&input_header_text, "h", "", "The input header if the stdin input has no header.")
	flag.StringVar(&input_comment, "c", "", "The input comment character, e.g., #.  Commented lines are not sent to output.")
	flag.IntVar(&input_reader_buffer_size, "input_reader_buffer_size", 4096, "The input reader buffer size") // default from https://golang.org/src/bufio/bufio.go

	flag.StringVar(&output_uri, "output_uri", "stdout", "The output uri")
	flag.StringVar(&output_format, "output_format", "", "The output format: "+strings.Join(GO_RAILGUN_FORMATS, ", "))

	flag.StringVar(&dfl_exp, "dfl_exp", "", "Process using dfl expression")
	flag.StringVar(&dfl_file, "dfl_file", "", "Process using dfl file.")

	flag.IntVar(&max_count, "max", -1, "The maximum number of objects to output")
	flag.BoolVar(&version, "version", false, "Prints version to stdout.")
	flag.BoolVar(&verbose, "verbose", false, "Prints verbose output.")
	flag.BoolVar(&help, "help", false, "Print help.")

	flag.Parse()

	if len(aws_default_region) == 0 {
		aws_default_region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if len(aws_access_key_id) == 0 {
		aws_access_key_id = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if len(aws_secret_access_key) == 0 {
		aws_secret_access_key = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if len(aws_session_token) == 0 {
		aws_session_token = os.Getenv("AWS_SESSION_TOKEN")
	}

	if len(hdfs_name_node) == 0 {
		hdfs_name_node = os.Getenv("HDFS_DEFAULT_NAME_NODE")
	}

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
		fmt.Println(railgun.VERSION)
		os.Exit(0)
	}

	if len(output_format) == 0 {
		fmt.Println("Error: Provided no -output_format.")
		fmt.Println("Run \"railgun -help\" for more information.")
		os.Exit(1)
	}

	var aws_session *session.Session
	var s3_client *s3.S3
	var hdfs_client *hdfs.Client

	if strings.HasPrefix(input_uri, "s3://") {
		aws_session = connect_to_aws(aws_access_key_id, aws_secret_access_key, aws_session_token, aws_default_region)
		s3_client = s3.New(aws_session)
	} else if strings.HasPrefix(input_uri, "hdfs://") {
		c, err := hdfs.New(hdfs_name_node)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error connecting to name node at uri "+hdfs_name_node))
		}
		hdfs_client = c
	}

	input_reader, input_metadata, err := reader.OpenResource(input_uri, input_compression, input_reader_buffer_size, false, s3_client, hdfs_client)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error opening resource from uri "+input_uri))
	}

	if len(input_format) == 0 {
		if input_metadata != nil {
			if len(input_metadata.ContentType) > 0 {
				switch input_metadata.ContentType {
				case "application/json":
					input_format = "json"
				case "application/vnd.geo+json":
					input_format = "json"
				case "application/toml":
					input_format = "toml"
				}
			}
			if len(input_format) == 0 {
				if strings.HasSuffix(input_uri, ".geojson") || strings.HasSuffix(input_uri, ".json") {
					input_format = "json"
				}
			}
		}
		if len(input_format) == 0 {
			fmt.Println("Error: Provided no -input_format and could not infer from resource.")
			fmt.Println("Run \"railgun -help\" for more information.")
			os.Exit(1)
		}
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

	funcs := dfl.NewFuntionMapWithDefaults()

	var dfl_node dfl.Node
	if len(dfl_file) > 0 {
		f, _, err := reader.OpenResource(dfl_file, "none", 4096, false, nil, nil)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error opening dfl file"))
		}
		content, err := f.ReadAll()
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error reading all from dfl file"))
		}
		dfl_exp = strings.TrimSpace(dfl.RemoveComments(string(content)))
	}
	if len(dfl_exp) > 0 {
		n, err := dfl.Parse(dfl_exp)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error parsing dfl node."))
		}
		dfl_node = n.Compile()
	}

	if verbose && dfl_node != nil {
		dfl_node_yaml, err := gss.Serialize(dfl_node.Map(), "yaml")
		if err != nil {
			log.Fatal(errors.Wrap(err, "error dumping dfl_node as yaml to stdout"))
		}
		fmt.Println(dfl_node_yaml)
	}

	input_object, _ := gss.NewObject(input_string, input_format)
	switch input_object_typed := input_object.(type) {
	case []map[string]interface{}:
		err = gss.Deserialize(input_string, input_format, input_header, input_comment, &input_object_typed)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error deserializing input using format "+input_format))
		}
		input_object = input_object_typed // This is a critical line, otherwise the type information is lost.
	default:
		err = gss.Deserialize(input_string, input_format, input_header, input_comment, &input_object)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error deserializing input using format "+input_format))
		}
	}

	var output interface{}
	if dfl_node != nil {
		o, err := dfl_node.Evaluate(input_object, funcs)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error processing"))
		}
		output = o
	} else {
		output = input_object
	}

	output = gss.StringifyMapKeys(output)

	output_string, err := gss.Serialize(output, output_format)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error converting"))
	}

	if output_uri == "stdout" {
		fmt.Println(output_string)
	} else if output_uri == "stderr" {
		fmt.Fprintf(os.Stderr, output_string)
	} else {
		output_file, err := os.OpenFile(output_uri, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(errors.Wrap(err, "error opening output file"))
		}
		w := bufio.NewWriter(output_file)
		_, err = w.WriteString(output_string + "\n")
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error writing string to output file"))
		}
		err = w.Flush()
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error flushing output to output file"))
		}
		err = output_file.Close()
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error closing output file."))
		}
	}

}
