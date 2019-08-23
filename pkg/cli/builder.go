// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"fmt"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-pipe/pkg/pipe"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/jsonl"
	"github.com/spatialcurrent/go-simple-serializer/pkg/sv"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

func BuildWriter(outputUri string, outputCompression string, outputFormat string, outputAppend bool, outputHeader []interface{}, outputKeySerializer stringify.Stringer, outputValueSerializer stringify.Stringer, outputNewLine string, s3Client *s3.S3) (pipe.Writer, error) {
	outputWriter, err := grw.WriteToResource(outputUri, outputCompression, outputAppend, s3Client)
	if err != nil {
		return nil, errors.Wrap(err, "error opening output file")
	}
	if outputFormat == "csv" || outputFormat == "tsv" {
		separator, err := sv.FormatToSeparator(outputFormat)
		if err != nil {
			return nil, err
		}
		return sv.NewWriter(outputWriter, separator, outputHeader, outputKeySerializer, outputValueSerializer, true, false), nil
	} else if outputFormat == "jsonl" {
		return jsonl.NewWriter(outputWriter, outputNewLine, outputKeySerializer, false), nil
	}
	return nil, errors.New(fmt.Sprintf("output format %q is no currently suported for streaming", outputFormat))
}
