// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-pipe/pkg/pipe"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-reader-writer/pkg/os"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-simple-serializer/pkg/iterator"
	gsswriter "github.com/spatialcurrent/go-simple-serializer/pkg/writer"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"

	"github.com/spatialcurrent/railgun/pkg/config"
	"github.com/spatialcurrent/railgun/pkg/util"
	railgunwriter "github.com/spatialcurrent/railgun/pkg/writer"
)

type ProcessAsStreamInput struct {
	Reader   io.Reader
	Type     reflect.Type
	Config   *config.Process
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func ProcessAsStream(input *ProcessAsStreamInput) error {

	logger := input.Logger

	if input.Config.Verbose {
		logger.Info("Processing as stream.")
		logger.Flush()
	}

	if !(input.Config.Output.CanStream()) {
		return fmt.Errorf("output format %q is not compatible with streaming", input.Config.Output.Format)
	}

	if input.Config.Output.IsEncrypted() {
		return errors.New("output passphrase is not compatible with streaming because it uses a block cipher")
	}

	//if !CanStream(input.Config.Input.Format, input.Config.Output.Format, input.Config.Output.Sorted) {
	if input.Config.Output.Sorted || !input.Config.Output.CanStream() {
		return errors.New("cannot stream with these inputs")
	}

	p := pipe.NewBuilder().InputLimit(input.Config.Input.Limit).OutputLimit(input.Config.Output.Limit)

	//
	// Input
	//

	if input.Config.Input.IsEncrypted() || !(input.Config.Input.CanStream()) {
		//
		// Sink
		//
		p = p.Input(NewSink(func() (interface{}, error) {
			inputBytes, err := util.DecryptReader(input.Reader, input.Config.Input.Passphrase, input.Config.Input.Salt)
			if err != nil {
				return nil, errors.Wrap(err, "error decoding input")
			}

			inputType, err := gss.GetType(inputBytes, input.Config.Input.Format)
			if err != nil {
				return nil, errors.Wrap(err, "error getting type for input")
			}

			if !(inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice) {
				return nil, errors.New("input type cannot be streamed as it is not an array or slice but " + fmt.Sprint(inputType))
			}

			inputObjects, err := gss.DeserializeBytes(&gss.DeserializeBytesInput{
				Bytes:         inputBytes,
				Format:        input.Config.Input.Format,
				Header:        input.Config.Input.Header,
				Comment:       input.Config.Input.Comment,
				LazyQuotes:    input.Config.Input.LazyQuotes,
				SkipLines:     input.Config.Input.SkipLines,
				Limit:         input.Config.Input.Limit,
				Type:          inputType,
				LineSeparator: input.Config.Input.LineSeparator,
			})
			if err != nil {
				return nil, errors.Wrapf(err, "error deserializing input using format %q", input.Config.Input.Format)
			}
			return inputObjects, nil
		}))
	} else if input.Config.Input.IsAthenaStoredQuery() {
		//
		// Athena
		//
		/*
			it, err := ProcessAthenaInput(
				input.Config.Input.Uri,
				input.Config.Input.Limit,
				input.Config.Temp.Uri,
				input.Config.Output.Format,
				athenaClient,
				logger,
				input.Config.Verbose)
			if err != nil {
				return errors.Wrap(err, "error processing athena input")
			}
			p = p.Input(it)
		*/
		return errors.New("athena support not ready")
	} else {
		//
		// Iterator
		//
		var inputType reflect.Type
		if input.Type != nil {
			inputType = input.Type.Elem()
		}
		it, err := iterator.NewIterator(&iterator.NewIteratorInput{
			Reader:            input.Reader,
			Type:              inputType,
			Format:            input.Config.Input.Format,
			Header:            input.Config.Input.Header,
			SkipLines:         input.Config.Input.SkipLines,
			SkipBlanks:        true,
			SkipComments:      true,
			Comment:           input.Config.Input.Comment,
			Trim:              true,
			LazyQuotes:        input.Config.Input.LazyQuotes,
			Limit:             input.Config.Input.Limit,
			LineSeparator:     input.Config.Input.LineSeparator,
			KeyValueSeparator: input.Config.Input.KeyValueSeparator,
			DropCR:            input.Config.Input.DropCR,
		})
		if err != nil {
			if err == io.EOF {
				return io.EOF
			}
			return errors.Wrap(err, "error creating iterator")
		}
		p = p.Input(it)
	}

	//
	// Transform
	//

	// dflVars is also used by dynamic output uris
	dflVars, err := input.Config.Dfl.Variables()
	if err != nil {
		return errors.Wrap(err, "error getting DFL variable from process config")
	}

	dflNode, err := input.Config.Dfl.Node()
	if err != nil {
		return errors.Wrap(err, "error parsing DFL node")
	}

	if dflNode != nil {
		p = p.Transform(NewTransformFunction(dflNode, dflVars))
	}

	//
	// Output
	//

	p = p.CloseOutput(true)

	if outputDevice := os.OpenDevice(input.Config.Output.Uri); outputDevice != nil {
		//
		// Output Device (e.g, Stdout, Stderr)
		//
		w, err := railgunwriter.NewWriter(&railgunwriter.NewWriterInput{
			Writer:            outputDevice,
			Algorithm:         input.Config.Output.Compression,
			Dictionary:        input.Config.Output.Dictionary,
			Format:            input.Config.Output.Format,
			FormatSpecifier:   input.Config.Output.FormatSpecifier,
			Header:            input.Config.Output.Header,
			ExpandHeader:      input.Config.Output.ExpandHeader,
			KeySerializer:     input.Config.Output.KeySerializer(),
			ValueSerializer:   input.Config.Output.ValueSerializer(),
			KeyValueSeparator: input.Config.Output.KeyValueSeparator,
			LineSeparator:     input.Config.Output.LineSeparator,
			Fit:               input.Config.Output.Fit,
			Pretty:            input.Config.Output.Pretty,
			Sorted:            input.Config.Output.Sorted,
			Reversed:          input.Config.Output.Reversed,
		})
		if err != nil {
			return errors.Wrap(err, "error creating new formatted writer")
		}
		err = p.Output(w).Run()
		if err != nil {
			return errors.Wrapf(err, "error processing iterator to writer (%q)", input.Config.Output.Uri)
		}
		return nil
	}

	outputNode, err := dfl.ParseCompile(input.Config.Output.Uri)
	if err != nil {
		return errors.Wrapf(err, "error parsing output uri %q", input.Config.Output.Uri)
	}

	if outputLiteral, ok := outputNode.(*dfl.Literal); ok {
		//
		// Literal Output Uri
		//
		compressedWriter, err := grw.WriteToResource(&grw.WriteToResourceInput{
			Uri:      fmt.Sprint(outputLiteral.Value),
			Alg:      input.Config.Output.Compression,
			Dict:     input.Config.Output.Dictionary,
			Append:   input.Config.Output.Append,
			Parents:  input.Config.Output.Mkdirs,
			S3Client: input.S3Client,
		})
		if err != nil {
			return errors.Wrapf(err, "error creating writer for uri %q", fmt.Sprint(outputLiteral.Value))
		}
		formattedWriter, err := gsswriter.NewWriter(&gsswriter.NewWriterInput{
			Writer:            compressedWriter,
			Format:            input.Config.Output.Format,
			Header:            input.Config.Output.Header,
			ExpandHeader:      input.Config.Output.ExpandHeader,
			KeySerializer:     input.Config.Output.KeySerializer(),
			ValueSerializer:   input.Config.Output.ValueSerializer(),
			KeyValueSeparator: input.Config.Output.KeyValueSeparator,
			LineSeparator:     input.Config.Output.LineSeparator,
			Fit:               input.Config.Output.Fit,
			Pretty:            input.Config.Output.Pretty,
			Sorted:            input.Config.Output.Sorted,
			Reversed:          input.Config.Output.Reversed,
		})
		if err != nil {
			return errors.Wrap(err, "error creating new formatted writer")
		}
		err = p.Output(formattedWriter).Run()
		if err != nil {
			return errors.Wrap(err, "error processing iterator to writer")
		}
		return nil
	}

	//
	// Dynamic Output Uri
	//

	outputPathBuffersMutex := &sync.RWMutex{}
	outputPathWriters := map[string]pipe.Writer{}
	outputPathBuffers := map[string]io.Buffer{}

	p = p.OutputF(
		NewOutputFunction(&NewOutputFunctionInput{
			Node:              outputNode,
			Vars:              dflVars,
			Algorithm:         input.Config.Output.Compression,
			Dictionary:        input.Config.Output.Dictionary,
			Format:            input.Config.Output.Format,
			Header:            input.Config.Output.Header,
			ExpandHeader:      input.Config.Output.ExpandHeader,
			KeySerializer:     input.Config.Output.KeySerializer(),
			ValueSerializer:   input.Config.Output.ValueSerializer(),
			KeyValueSeparator: input.Config.Output.KeyValueSeparator,
			LineSeparator:     input.Config.Output.LineSeparator,
			Fit:               input.Config.Output.Fit,
			Pretty:            input.Config.Output.Pretty,
			Sorted:            input.Config.Output.Sorted,
			Reversed:          input.Config.Output.Reversed,
			Mutex:             outputPathBuffersMutex,
			PathWriters:       outputPathWriters,
			PathBuffers:       outputPathBuffers,
		}),
	)

	logger.Info(map[string]interface{}{
		"msg": "starting pipeline",
	})
	logger.Flush()

	err = p.Run()
	if err != nil {
		return errors.Wrap(err, "error processing iterator to writer")
	}

	logger.Info(map[string]interface{}{
		"msg":               "done running pipeline",
		"outputPathWriters": len(outputPathWriters),
		"outputPathBuffers": len(outputPathBuffers),
	})
	logger.Flush()

	for _, w := range outputPathWriters {

		err := io.Flush(w)
		if err != nil {
			return errors.Wrap(err, "error flushing output writer")
		}

		err = io.Close(w)
		if err != nil {
			return errors.Wrap(err, "error closing output writer")
		}
	}

	err = grw.WriteBuffers(&grw.WriteBuffersInput{
		Buffers:    outputPathBuffers,
		Algorithm:  "none",
		Dictionary: grw.NoDict,
		Overwrite:  input.Config.Output.Overwrite,
		Append:     input.Config.Output.Append,
		Mkdirs:     input.Config.Output.Mkdirs,
		S3Client:   input.S3Client,
	})
	if err != nil {
		return errors.Wrap(err, "error writing buffers to files")
	}

	return nil
}
