// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

func coalesce(strs ...string) string {
	for _, str := range strs {
		if len(str) > 0 {
			return str
		}
	}
	return ""
}
