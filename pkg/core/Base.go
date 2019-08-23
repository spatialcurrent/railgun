// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
)

type Base interface {
	Named
	Mapper
	Dfl(ctx context.Context) string
}
