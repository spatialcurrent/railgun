// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
)

type Mapper interface {
	Map(ctx context.Context) map[string]interface{}
}
