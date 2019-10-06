// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serializer

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"

	"github.com/spatialcurrent/railgun/pkg/config"
)

type WriteObjectInput struct {
	Reader   io.Reader
	Type     reflect.Type
	Config   *config.Process
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func WriteObject(input *WriteObjectInput) error {

	outputBytes, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            outputObject,
		Format:            input.Config.Output.Format,
		FormatSpecifier:   input.Config.Output.FormatSpecifier,
		Header:            input.Config.Output.Header,
		Limit:             input.Config.Output.Limit,
		Pretty:            input.Config.Output.Pretty,
		Sorted:            input.Config.Output.Sorted,
		Reversed:          input.Config.Output.Reversed,
		LineSeparator:     input.Config.Output.LineSeparator,
		KeyValueSeparator: input.Config.Output.KeyValueSeparator,
		KeySerializer:     input.Config.Output.KeySerializer(),
		ValueSerializer:   input.Config.Output.ValueSerializer(),
		EscapePrefix:      input.Config.Output.EscapePrefix,
		EscapeSpace:       input.Config.Output.EscapeSpace,
		EscapeNewLine:     input.Config.Output.EscapeNewLine,
		EscapeEqual:       input.Config.Output.EscapeEqual,
		EscapeColon:       input.Config.Output.EscapeColon,
		ExpandHeader:      input.Config.Output.ExpandHeader,
	})

	if err != nil {
		return errors.Wrap(err, "error evaluating filter")
	}

	outputWriter, err := grw.WriteToResource(&grw.WriteToResourceInput{
		Uri:      input.Config.Output.Uri,
		Alg:      input.Config.Output.Compression,
		Dict:     grw.NoDict,
		Append:   input.Config.Output.Append,
		S3Client: input.S3Client,
	})
	if err != nil {
		return errors.Wrap(err, "error opening output file")
	}

	_, err = outputWriter.Write(outputBytes)
	if err != nil {
		return errors.Wrap(err, "error writing to output file")
	}

	err = outputWriter.Flush()
	if err != nil {
		return errors.Wrap(err, "error flushing to output file")
	}

	err = outputWriter.Close()
	if err != nil {
		return errors.Wrap(err, "error writing to output file")
	}

	return nil
}
