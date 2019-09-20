// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package jwt

import (
	"fmt"
	"time"
)

type ErrInvalidSessionDuration struct {
	Value time.Duration
	Min   time.Duration
}

func (e *ErrInvalidSessionDuration) Error() string {
	return fmt.Sprintf("invalid session duration %v, must be greater than or equal to %v", e.Value, e.Min)
}
