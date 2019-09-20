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

type Process struct {
	Name        string                 `map:"name" rest:"name, the unique name of the process" required:"yes"`
	Title       string                 `map:"title,omitempty" rest:"title, the title of the process"`
	Description string                 `map:"description,omitempty" rest:"description, a verbose description of the process"`
	Node        dfl.Node               `map:"expression,omitempty" rest:"expression, the DFL expression of the process" required:"yes"`
	Defaults    map[string]interface{} `map:"defaults,omitempty" rest:"defaults, the default values of the variables for this process"`
	Tags        []string               `map:"tags,omitempty" rest:"tags, tags for the service"`
}

func (p Process) GetName() string {
	return p.Name
}

func (p Process) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, p)
}

func (p Process) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range p.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var ProcessType = reflect.TypeOf(Process{})
