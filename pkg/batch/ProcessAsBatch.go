// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package batch

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
	"github.com/spatialcurrent/railgun/pkg/serializer"
)

type ProcessAsBatchInput struct {
	Reader   io.Reader
	Type     reflect.Type
	Config   *config.Process
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func ProcessAsBatch(input *ProcessAsBatchInput) error {

	inputObject, err := gss.DeserializeReader(&gss.DeserializeReaderInput{
		Reader:          input.Reader,
		Type:            input.Type,
		Format:          input.Config.Input.Format,
		Header:          input.Config.Input.Header,
		Comment:         input.Config.Input.Comment,
		LazyQuotes:      input.Config.Input.LazyQuotes,
		SkipLines:       input.Config.Input.SkipLines,
		SkipBlanks:      input.Config.Input.SkipBlanks,
		SkipComments:    input.Config.Input.SkipComments,
		Trim:            input.Config.Input.Trim,
		Limit:           input.Config.Input.Limit,
		LineSeparator:   input.Config.Input.LineSeparator,
		DropCR:          input.Config.Input.DropCR,
		EscapePrefix:    input.Config.Input.EscapePrefix,
		UnescapeSpace:   input.Config.Input.UnescapeSpace,
		UnescapeNewLine: input.Config.Input.UnescapeNewLine,
		UnescapeColon:   input.Config.Input.UnescapeColon,
		UnescapeEqual:   input.Config.Input.UnescapeEqual,
	})
	if err != nil {
		return errors.Wrap(err, "error deserializing input object")
	}

	dflNode, err := input.Config.Dfl.Node()
	if err != nil {
		return errors.Wrap(err, "error parsing dfl node")
	}

	var outputObject interface{}
	if dflNode != nil {

		dflVars, err := input.Config.Dfl.Variables()
		if err != nil {
			return errors.Wrap(err, "error getting dfl variables")
		}

		_, object, err := dflNode.Evaluate(dflVars, inputObject, dfl.DefaultFunctionMap, []string{"'", "\"", "`"})
		if err != nil {
			return errors.Wrap(err, "error evaluating filter")
		}

		outputObject = object
	} else {
		outputObject = inputObject
	}

	err = serializer.Serialize(&gss.SerializeInput{
		Uri:               input.Config.Output.Uri,
		Alg:               input.Config.Output.Compression,
		Dict:              grw.NoDict,
		Append:            input.Config.Output.Append,
		Parents:           input.Config.Output.Parents,
		S3Client:          input.S3Client,
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
		return errors.Wrapf(err, "error writing object to uri %q", input.Config.Output.Uri)
	}

	return nil
}
