// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package serializer

import (
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
)

type SerializeInput struct {
	Uri               string
	Alg               string
	Dict              []byte
	Append            bool
	Parents           bool
	S3Client          *s3.S3
	Object            interface{}
	Format            string
	FormatSpecifier   string
	Fit               bool
	Header            []interface{}
	Limit             int
	Pretty            bool
	Sorted            bool
	Reversed          bool
	LineSeparator     string
	KeyValueSeparator string
	KeySerializer     func(object interface{}) (string, error)
	ValueSerializer   func(object interface{}) (string, error)
	EscapePrefix      string
	EscapeSpace       bool
	EscapeNewLine     bool
	EscapeEqual       bool
	EscapeColon       bool
	ExpandHeader      bool
}

func Serialize(input *SerializeInput) error {
	b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
		Object:            input.Object,
		Format:            input.Format,
		FormatSpecifier:   input.FormatSpecifier,
		Header:            input.Header,
		Limit:             input.Limit,
		Pretty:            input.Pretty,
		Sorted:            input.Sorted,
		Reversed:          input.Reversed,
		LineSeparator:     input.LineSeparator,
		KeyValueSeparator: input.KeyValueSeparator,
		KeySerializer:     input.KeySerializer,
		ValueSerializer:   input.ValueSerializer,
		EscapePrefix:      input.EscapePrefix,
		EscapeSpace:       input.EscapeSpace,
		EscapeNewLine:     input.EscapeNewLine,
		EscapeEqual:       input.EscapeEqual,
		EscapeColon:       input.EscapeColon,
		ExpandHeader:      input.ExpandHeader,
	})
	if err != nil {
		return errors.Wrapf(err, "error serializing object for uri %q", input.Uri)
	}

	err = grw.WriteAllAndClose(&grw.WriteAllAndCloseInput{
		Bytes:    b,
		Uri:      input.Uri,
		Alg:      input.Alg,
		Dict:     grw.NoDict,
		Append:   input.Append,
		Parents:  input.Parents,
		S3Client: input.S3Client,
	})
	if err != nil {
		return errors.Wrapf(err, "error writing object to uri %q", input.Uri)
	}

	return nil
}
