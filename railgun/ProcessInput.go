// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"github.com/pkg/errors"
	//"github.com/spatialcurrent/go-reader/reader"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	//"os"
	//"strings"
)

func ProcessInput(inputBytes []byte, inputFormat string, inputHeader []string, inputComment string, inputLazyQuotes bool, inputLimit int, dflExpression string, dflUri string, outputFormat string, outputHeader []string, outputLimit int, verbose bool) (string, error) {

	if len(outputFormat) > 0 {

		output, err := ProcessObject(inputBytes, inputFormat, inputHeader, inputComment, inputLazyQuotes, inputLimit, dflExpression, dflUri, verbose)
		if err != nil {
			return "", errors.Wrap(err, "error processing object")
		}

		output = gss.StringifyMapKeys(output)

		outputString, err := gss.SerializeString(output, outputFormat, outputHeader, outputLimit)
		if err != nil {
			return "", errors.Wrap(err, "error converting output")
		}

		return outputString, nil
	}

	return string(inputBytes), nil

}
