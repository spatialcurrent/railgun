// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-pipe/pkg/pipe"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-simple-serializer/pkg/writer"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

type NewOutputFunctionInput struct {
	Node              dfl.Node // output node
	Vars              map[string]interface{}
	Format            string
	Algorithm         string
	Dictionary        []byte
	Header            []interface{}
	ExpandHeader      bool
	KeySerializer     stringify.Stringer
	ValueSerializer   stringify.Stringer
	LineSeparator     string
	KeyValueSeparator string
	Fit               bool
	Pretty            bool
	Sorted            bool
	Reversed          bool
	Mutex             *sync.RWMutex
	PathWriters       map[string]pipe.Writer
	PathBuffers       map[string]io.Buffer
}

func NewOutputFunction(input *NewOutputFunctionInput) func(object interface{}) error {
	return func(object interface{}) error {

		_, outputPath, err := input.Node.Evaluate(input.Vars, object, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			return errors.Wrap(err, "error evaluating filter")
		}

		outputPathString, ok := outputPath.(string)
		if !ok {
			return errors.New("output path is not a string")
		}

		input.Mutex.Lock()
		if _, ok := input.PathBuffers[outputPathString]; !ok {

			compressedWriter, outputBuffer, err := grw.WriteBytes(input.Algorithm, input.Dictionary)
			if err != nil {
				return errors.Wrapf(err, "error writing to bytes for compression %q", input.Algorithm)
			}

			formattedWriter, err := writer.NewWriter(&writer.NewWriterInput{
				Writer:            compressedWriter,
				Format:            input.Format,
				Header:            input.Header,
				ExpandHeader:      input.ExpandHeader,
				KeySerializer:     input.KeySerializer,
				ValueSerializer:   input.ValueSerializer,
				KeyValueSeparator: input.KeyValueSeparator,
				LineSeparator:     input.LineSeparator,
				Fit:               input.Fit,
				Pretty:            input.Pretty,
				Sorted:            input.Sorted,
				Reversed:          input.Reversed,
			})
			if err != nil {
				return errors.Wrap(err, "error creating formatted writer")
			}
			input.PathBuffers[outputPathString] = outputBuffer
			input.PathWriters[outputPathString] = formattedWriter
		}
		input.Mutex.Unlock()

		err = input.PathWriters[outputPathString].WriteObject(object)
		if err != nil {
			return errors.Wrap(err, "error writing object to buffer")
		}

		err = input.PathWriters[outputPathString].Flush()
		if err != nil {
			return errors.Wrap(err, "error flushing object to buffer")
		}

		return nil
	}
}
