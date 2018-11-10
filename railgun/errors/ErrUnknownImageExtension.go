// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package errors

type ErrUnknownImageExtension struct {
	Extension string
}

func (e *ErrUnknownImageExtension) Error() string {
	return "unknown image extension " + e.Extension
}
