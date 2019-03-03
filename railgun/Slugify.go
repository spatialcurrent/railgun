// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

func Slugify(s string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		panic(errors.Wrap(err, "invalid regular expression for slugify"))
	}
	return reg.ReplaceAllString(strings.ToLower(s), "-")
}
