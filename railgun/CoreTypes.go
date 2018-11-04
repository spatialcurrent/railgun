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
