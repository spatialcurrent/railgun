// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/go-try-get/pkg/gtg"
)

import (
	"github.com/spatialcurrent/railgun/pkg/cache"
	"github.com/spatialcurrent/railgun/pkg/core"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/parser"
	"github.com/spatialcurrent/railgun/pkg/util"
)


type SaveToUriInput struct {
  Uri string
  Format string
  Compression string
  Logger *gsl.Logger
  S3Client *s3.S3
}

func (c *RailgunCatalog) SaveToUri(input *SaveToUriInput) error {

	err := func(data map[string]interface{}) error {

		b, err := gss.SerializeBytes(&gss.SerializeBytesInput{
			Object:            data,
			Format:            input.Format,
			Header:            gss.NoHeader,
			Limit:             gss.NoLimit,
			Pretty:            false,
			LineSeparator:     "\n",
			KeyValueSeparator: "=",
		})
		if err != nil {
			return errors.Wrap(err, "error serializing catalog")
		}

		if scheme == "s3" {
			i := strings.Index(uriPath, "/")
			if i == -1 {
				return errors.New("s3 path missing bucket")
			}
			err := grw.UploadS3Object(uriPath[0:i], uriPath[i+1:], bytes.NewBuffer(b), s3Client)
			if err != nil {
				return errors.Wrap(err, "error uploading new version of catalog to S3")
			}
			return nil
		}

		outputWriter, err := grw.WriteAllAndClose(&grw.WriteToResourceInput{
		  Bytes: b,
			Uri:      input.Uri,
			Alg:      input.Compression,
			Dict:     grw.NoDict,
			Append:   false,
			Parents: false,
			S3Client: nil,
		})
		if err != nil {
			return errors.Wrap(err, "error writing writer")
		}

		return nil

	}(c.Dump(nil))

	if err != nil {
		return errors.Wrap(err, "error saving catalog")
	}

	return nil
}
