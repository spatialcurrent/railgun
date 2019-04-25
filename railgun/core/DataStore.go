// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"context"
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type DataStore struct {
	Workspace   *Workspace             `rest:"workspace, the name of the containing workspace" required:"yes" visibility:"public"`
	Name        string                 `rest:"name, the unique name of the data store" required:"yes" visibility:"public"`
	Title       string                 `rest:"title, the title of the data store" visibility:"public"`
	Description string                 `rest:"description, a verbose description of the data store" visibility:"public"`
	Uri         dfl.Node               `rest:"uri, a uri to the data (local or AWS s3)" required:"yes" visibility:"private"`
	Format      string                 `rest:"format, the format of the data (default inferred from uri)" visibility:"private"`
	Compression string                 `rest:"compression, the compression of the data (default inferred from uri)" visibility:"private"`
	Extent      []float64              `rest:"extent, the extent of the data" visibility:"public"`
	Vars        map[string]interface{} `rest:"vars, the values of the variables for this data store"`
	Filter      dfl.Node               `rest:"filter, a filter to apply when querying this datastore" required:"no" visibility:"private"`
}

func (ds DataStore) GetName() string {
	return ds.Name
}

func (ds DataStore) Map(ctx context.Context) map[string]interface{} {
	m := map[string]interface{}{
		"workspace":   ds.Workspace.Name,
		"name":        ds.Name,
		"title":       ds.Title,
		"description": ds.Description,
		"uri":         ds.Uri.Dfl(dfl.DefaultQuotes, false, 0),
		"format":      ds.Format,
		"compression": ds.Compression,
		"extent":      dfl.Literal{Value: ds.Extent}.Dfl(dfl.DefaultQuotes, false, 0),
	}
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ds.Vars {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	if len(dict) > 0 {
		m["vars"] = dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
	}
	if ds.Filter != nil {
		m["filter"] = ds.Filter.Dfl(dfl.DefaultQuotes, false, 0)
	}
	return m
}

func (ds DataStore) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ds.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var DataStoreType = reflect.TypeOf(DataStore{})
