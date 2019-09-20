// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package http

import (
	"fmt"
	"time"
)

type ErrInvalidTimeoutWrite struct {
	Value time.Duration
	Min   time.Duration
}

func (e *ErrInvalidTimeoutWrite) Error() string {
	return fmt.Sprintf("invalid write timeout %v, must be greater than or equal to %v", e.Value, e.Min)
}
