// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

// Package iterator provides an easy API to create an iterator to write compressed objects to a file.
package writer

import (
	"io"

	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-pipe/pkg/pipe"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/writer"
	"github.com/spatialcurrent/go-stringify/pkg/stringify"
)

// Parameters for NewWriter function.
type NewWriterInput struct {
	Writer            io.Writer
	Algorithm         string
	Dictionary        []byte
	Format            string
	FormatSpecifier   string
	Header            []interface{}
	ExpandHeader      bool // in context, only used by tags as ExpandKeys
	KeySerializer     stringify.Stringer
	ValueSerializer   stringify.Stringer
	KeyValueSeparator string
	LineSeparator     string
	Fit               bool
	Pretty            bool
	Sorted            bool
	Reversed          bool
}

// NewWriter returns a new pipe.Writer for writing compressed and formatted objects to an underlying writer.
func NewWriter(input *NewWriterInput) (pipe.Writer, error) {

	compressedWriter, err := grw.WrapWriter(input.Writer, input.Algorithm, input.Dictionary)
	if err != nil {
		return nil, errors.Wrap(err, "error creating compressed writer")
	}

	formattedWriter, err := writer.NewWriter(&writer.NewWriterInput{
		Writer:            compressedWriter,
		Format:            input.Format,
		FormatSpecifier:   input.FormatSpecifier,
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
		return nil, errors.Wrap(err, "error creating new formatted writer")
	}
	return formattedWriter, nil
}
