// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package stream

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-pipe/pkg/pipe"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/io"
	"github.com/spatialcurrent/go-simple-serializer/pkg/jsonl"
	"github.com/spatialcurrent/go-simple-serializer/pkg/sv"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

type NewOutputFunctionInput struct {
	Node            dfl.Node // output node
	Vars            map[string]interface{}
	Format          string
	Compression     string
	Header          []interface{}
	KeySerializer   stringify.Stringer
	ValueSerializer stringify.Stringer
	LineSeparator   string
	Mutex           *sync.RWMutex
	PathWriters     map[string]pipe.Writer
	PathBuffers     map[string]io.Buffer
}

//func NewOutputFunction(outputNode dfl.Node, dflVars map[string]interface{}, outputFormat string, outputCompression string, outputHeader []interface{}, outputKeySerializer stringify.Stringer, outputValueSerializer stringify.Stringer, outputLineSeparator string, outputPathBuffersMutex *sync.RWMutex, outputPathWriters map[string]pipe.Writer, outputPathBuffers map[string]grw.Buffer) func(object interface{}) error {

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

			outputWriter, outputBuffer, err := grw.WriteBytes(input.Compression, []byte{})
			if err != nil {
				return errors.Wrapf(err, "error writing to bytes for compression %q", input.Compression)
			}

			if input.Format == "csv" || input.Format == "tsv" {
				separator, err := sv.FormatToSeparator(input.Format)
				if err != nil {
					return err
				}
				input.PathWriters[outputPathString] = sv.NewWriter(
					outputWriter,
					separator,
					input.Header,
					input.KeySerializer,
					input.ValueSerializer,
					true,
					false)
			} else if input.Format == "jsonl" {
				input.PathWriters[outputPathString] = jsonl.NewWriter(outputWriter, input.LineSeparator, input.KeySerializer, false)
			} else {
				return fmt.Errorf("cannot create streaming writer for format %q", input.Format)
			}

			input.PathBuffers[outputPathString] = outputBuffer
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
