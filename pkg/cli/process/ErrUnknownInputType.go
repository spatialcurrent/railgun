// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package process

import (
	"fmt"
)

type ErrUnknownInputType struct {
	Value string
}

func (e *ErrUnknownInputType) Error() string {
	return fmt.Sprintf("unknown input type %q", e.Value)
}
