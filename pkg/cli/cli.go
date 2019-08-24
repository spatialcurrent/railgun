// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package cli

import (
	"github.com/pkg/errors"
)

var (
	ErrMissingLineSeparator     = errors.New("line separator cannot be blank")
	ErrMissingKeyValueSeparator = errors.New("key-value separator cannot be blank")
	ErrMissingEscapePrefix      = errors.New("escape prefix is missing")
)
