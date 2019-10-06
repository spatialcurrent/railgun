// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"fmt"
	"strings"
)

type ErrInvalidCatalogCompression struct {
	Value    string
	Expected []string
}

func (e *ErrInvalidCatalogCompression) Error() string {
	return fmt.Sprintf("invalid catalog compression %q, expecting on of %s", e.Value, strings.Join(e.Expected, ", "))
}
