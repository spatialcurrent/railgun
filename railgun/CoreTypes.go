// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"reflect"
)

var CoreTypes = map[string]reflect.Type{
	"workspace": WorkspaceType,
	"datastore": DataStoreType,
	"layer":     LayerType,
	"process":   ProcessType,
	"service":   ServiceType,
	"job":       JobType,
}
