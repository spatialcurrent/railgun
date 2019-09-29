// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

/*

import (
	"fmt"
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

type ProcessAsStreamToStreamInput struct {
	Reader   io.Reader
	Type     reflect.Type
	Config   *config.Process
	DflVars  map[string]interface{}
	DflNode  dfl.Node
	S3Client *s3.S3
	Logger   *gsl.Logger
}

func ProcessAsStreamToStream(input *ProcessAsStreamToStreamInput) error {
	if processConfig.Verbose {
		logger.Info("Processing as stream to stream.")
		logger.Flush()
	}

	p := pipe.NewBuilder().OutputLimit(processConfig.Output.Limit).Filter(pipe.FilterNotNil)

	if len(processConfig.Input.LineSeparator) != 1 {
		return fmt.Errorf("invalid line separator %q with length %d", processConfig.Input.LineSeparator, len(processConfig.Input.LineSeparator))
	}

	iteratorInput := &iterator.NewIteratorInput{
		Reader:            inputReader,
		Type:              input.Type,
		Format:            processConfig.Input.Format,
		Header:            processConfig.Input.Header,
		SkipLines:         processConfig.Input.SkipLines,
		SkipBlanks:        true,
		SkipComments:      true,
		Comment:           processConfig.Input.Comment,
		Trim:              true,
		LazyQuotes:        processConfig.Input.LazyQuotes,
		Limit:             processConfig.Input.Limit,
		LineSeparator:     processConfig.Input.LineSeparator,
		KeyValueSeparator: processConfig.Input.KeyValueSeparator,
		DropCR:            processConfig.Input.DropCR,
	}

	it, err := iterator.NewIterator(iteratorInput)
	if err != nil {
		if err == io.EOF {
			return io.EOF
		}
		return errors.Wrap(err, "error creating iterator")
	}
	p = p.Input(it)

	if dflNode != nil {
		p = p.Transform(stream.NewTransformFunction(dflNode, dflVars))
	}

	outputUri := processConfig.Output.Uri
	outputFormat := processConfig.Output.Format
	outputCompression := processConfig.Output.Compression
	outputHeader := make([]interface{}, 0)
	for _, str := range processConfig.Output.Header {
		outputHeader = append(outputHeader, str)
	}
	outputAppend := processConfig.Output.Append
	outputOverwrite := processConfig.Output.Overwrite
	outputMkdirs := processConfig.Output.Mkdirs
	outputKeySerializer := processConfig.Output.KeySerializer()
	outputValueSerializer := processConfig.Output.ValueSerializer()

	outputLineSeparator := processConfig.Output.LineSeparator
	outputSorted := processConfig.Output.Sorted
	outputReversed := processConfig.Output.Reversed
	outputPretty := processConfig.Output.Pretty

	if outputDevice := os.OpenDevice(outputUri); outputDevice != nil {
		switch outputFormat {
		case "csv", "tsv":
			separator, err := sv.FormatToSeparator(outputFormat)
			if err != nil {
				return err
			}
			p = p.Output(
				sv.NewWriter(
					outputDevice,
					separator,
					outputHeader,
					outputKeySerializer,
					outputValueSerializer,
					outputSorted,
					outputReversed,
				),
			)
		case "jsonl":
			p = p.Output(
				jsonl.NewWriter(
					outputDevice,
					outputLineSeparator,
					outputKeySerializer,
					outputPretty,
				),
			)
		}
		err = p.Run()
		if err != nil {
			return errors.Wrap(err, "error processing iterator to writer")
		}
	} else {

		outputNode, err := dfl.ParseCompile(outputUri)
		if err != nil {
			return errors.Wrap(err, "error parsing output uri: "+outputUri)
		}

		outputPathBuffersMutex := &sync.RWMutex{}
		outputPathWriters := map[string]pipe.Writer{}
		outputPathBuffers := map[string]io.Buffer{}

		p = p.OutputF(
			stream.NewOutputFunction(&stream.NewOutputFunctionInput{
				Node:            outputNode,
				Vars:            dflVars,
				Format:          outputFormat,
				Compression:     outputCompression,
				Header:          outputHeader,
				KeySerializer:   outputKeySerializer,
				ValueSerializer: outputValueSerializer,
				LineSeparator:   outputLineSeparator,
				Mutex:           outputPathBuffersMutex,
				PathWriters:     outputPathWriters,
				PathBuffers:     outputPathBuffers,
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
				return errors.Wrap(err, "error flushing output writer")
			}
		}

		err = grw.WriteBuffers(&grw.WriteBuffersInput{
			Buffers:   outputPathBuffers,
			Algorithm: "none",
			Overwrite: outputOverwrite,
			Append:    outputAppend,
			Mkdirs:    outputMkdirs,
			S3Client:  s3Client,
		})
		if err != nil {
			return errors.Wrap(err, "error writing buffers to files")
		}
	}

	if processConfig.Time {
		logger.Info(map[string]interface{}{
			"msg": "ended",
		})
	}

	return nil
}

*/
