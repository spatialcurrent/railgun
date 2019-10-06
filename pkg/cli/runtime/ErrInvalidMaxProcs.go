// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package runtime

import (
	"fmt"
)

type ErrInvalidMaxProcs struct {
	Value int
}

func (e *ErrInvalidMaxProcs) Error() string {
	return fmt.Sprintf("invalid max procs %d, must be greater than or equal to 0", e.Value)
}
