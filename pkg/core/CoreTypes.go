// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"reflect"
)

var CoreTypes = map[string]reflect.Type{
	"workspace": WorkspaceType,
	"datastore": DataStoreType,
	"layer":     LayerType,
	"function":  FunctionType,
	"process":   ProcessType,
	"service":   ServiceType,
	"job":       JobType,
	"workflow":  WorkflowType,
	"user":      UserType,
}
