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

type Service struct {
	Name        string                 `map:"name" rest:"name, the unique name of the service" required:"yes"`
	Title       string                 `map:"title" rest:"title, the title of the service"`
	Description string                 `map:"description" rest:"description, a verbose description of the service"`
	DataStore   *DataStore             `map:"datastore" rest:"datastore, the name of the data store" required:"yes"`
	Process     *Process               `map:"process" rest:"process, the name of the process" required:"yes"`
	Defaults    map[string]interface{} `map:"defaults,omitempty" rest:"defaults, the default values of the variables for this service"`
	Tags        []string               `map:"tags,omitempty" rest:"tags, tags for the service"`
	Transform   dfl.Node               `map:"transform,omitempty" rest:"transform, transform applied to request variables before passing to datastore and process" required:"no" visibility:"private"`
}

func (s Service) GetName() string {
	return s.Name
}

func (s Service) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, s)
}

func (s Service) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range s.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ServiceType = reflect.TypeOf(Service{})
