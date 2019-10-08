// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/railgun/pkg/serializer"
)

type SaveToUriInput struct {
	Uri         string
	Format      string
	Compression string
	S3Client    *s3.S3
}

func (c *RailgunCatalog) SaveToUri(input *SaveToUriInput) error {
	err := serializer.New(input.Format, input.Compression, grw.NoDict).
		S3Client(input.S3Client).
		Serialize(input.Uri, c.Dump(nil))
	if err != nil {
		return errors.Wrapf(err, "error saving catalog to uri %q", input.Uri)
	}

	return nil
}
