// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"reflect"

	"github.com/spatialcurrent/go-dfl/pkg/dfl"

	"github.com/spatialcurrent/railgun/pkg/mapper"
)

type DataStore struct {
	Workspace   *Workspace             `map:"workspace" rest:"workspace, the name of the containing workspace" required:"yes" visibility:"public"`
	Name        string                 `map:"name" rest:"name, the unique name of the data store" required:"yes" visibility:"public"`
	Title       string                 `map:"title" rest:"title, the title of the data store" visibility:"public"`
	Description string                 `map:"description" rest:"description, a verbose description of the data store" visibility:"public"`
	Uri         dfl.Node               `map:"uri" rest:"uri, a uri to the data (local or AWS s3)" required:"yes" visibility:"private"`
	Format      string                 `map:"format" rest:"format, the format of the data (default inferred from uri)" visibility:"private"`
	Compression string                 `map:"compression" rest:"compression, the compression of the data (default inferred from uri)" visibility:"private"`
	Extent      []float64              `map:"extent" rest:"extent, the extent of the data" visibility:"public"`
	Vars        map[string]interface{} `map:"vars,omitempty" rest:"vars, the values of the variables for this data store"`
	Filter      dfl.Node               `map:"filter,omitempty" rest:"filter, a filter to apply when querying this datastore" required:"no" visibility:"private"`
}

func (ds DataStore) GetName() string {
	return ds.Name
}

func (ds DataStore) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, ds)
}

func (ds DataStore) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ds.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var DataStoreType = reflect.TypeOf(DataStore{})
