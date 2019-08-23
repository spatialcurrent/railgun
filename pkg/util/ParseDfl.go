// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package util

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
)

func ParseDfl(dflUri string, dflExpression string) (dfl.Node, error) {
	var dflNode dfl.Node

	if len(dflUri) > 0 {
		f, _, err := grw.ReadFromResource(dflUri, "none", 4096, nil)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening dfl file")
		}
		content, err := f.ReadAll()
		if err != nil {
			return nil, errors.Wrap(err, "Error reading all from dfl file")
		}
		dflExpression = strings.TrimSpace(dfl.RemoveComments(string(content)))
	}

	if len(dflExpression) > 0 {
		n, err := dfl.ParseCompile(dflExpression)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing dfl node.")
		}
		dflNode = n
	}

	return dflNode, nil
}
