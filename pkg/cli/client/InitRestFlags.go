// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package client

import (
	"reflect"
	"strings"

	"github.com/spatialcurrent/pflag"
)

// InitRestFlags initializes flags based on the rest struct tags.
func InitRestFlags(flag *pflag.FlagSet, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if str, ok := f.Tag.Lookup("rest"); ok && str != "" && str != "-" {
			//t := f.Type
			if strings.Contains(str, ",") {
				arr := strings.SplitN(str, ",", 2)
				flag.String(arr[0], "", arr[1])
			} else {
				flag.String(str, "", "")
			}
			// use dfl for arrays
			/*if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
				if strings.Contains(str, ",") {
					arr := strings.SplitN(str, ",", 2)
					flag.StringArray(arr[0], []string{}, arr[1])
				} else {
					flag.StringArray(str, []string{}, "")
				}
			} else {
				if strings.Contains(str, ",") {
					arr := strings.SplitN(str, ",", 2)
					flag.String(arr[0], "", arr[1])
				} else {
					flag.String(str, "", "")
				}
			}*/
		}
	}
}
