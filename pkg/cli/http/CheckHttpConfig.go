// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package http

import (
	"github.com/spatialcurrent/viper"
)

// CheckHttpConfig checks the http configuration.
func CheckHttpConfig(v *viper.Viper) error {
	schemes := v.GetStringSlice(FlagHttpSchemes)
	if len(schemes) == 0 {
		return ErrMissingSchemes
	}
	timeoutIdle := v.GetDuration(FlagHttpTimeoutIdle)
	if timeoutIdle.Nanoseconds() < MinIdleTimeout.Nanoseconds() {
		return &ErrInvalidTimeoutIdle{Value: timeoutIdle, Min: MinIdleTimeout}
	}
	timeoutRead := v.GetDuration(FlagHttpTimeoutRead)
	if timeoutRead.Nanoseconds() < MinReadTimeout.Nanoseconds() {
		return &ErrInvalidTimeoutRead{Value: timeoutRead, Min: MinReadTimeout}
	}
	timeoutWrite := v.GetDuration(FlagHttpTimeoutWrite)
	if timeoutWrite.Nanoseconds() < MinWriteTimeout.Nanoseconds() {
		return &ErrInvalidTimeoutWrite{Value: timeoutWrite, Min: MinWriteTimeout}
	}
	gracefulShutdownWait := v.GetDuration(FlagHttpGracefulShutdownWait)
	if gracefulShutdownWait.Nanoseconds() < MinGracefulShutdownWait.Nanoseconds() {
		return &ErrInvalidGracefulShutdownWait{Value: gracefulShutdownWait, Min: MinGracefulShutdownWait}
	}
	return nil
}
