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

type ErrInvalidCatalogFormat struct {
	Value    string
	Expected []string
}

func (e *ErrInvalidCatalogFormat) Error() string {
	return fmt.Sprintf("invalid catalog format %q, expecting on of %s", e.Value, strings.Join(e.Expected, ", "))
}
