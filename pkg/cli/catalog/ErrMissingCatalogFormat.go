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

type ErrMissingCatalogFormat struct {
	Expected []string
}

func (e *ErrMissingCatalogFormat) Error() string {
	return fmt.Sprintf("missing catalog format, expecting on of %s", strings.Join(e.Expected, ", "))
}
