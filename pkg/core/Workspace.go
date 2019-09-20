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

type Workspace struct {
	Name        string `map:"name" rest:"name, the unique name of the workspace" required:"yes"`
	Title       string `map:"title" rest:"title, the title of the workspace"`
	Description string `map:"description" rest:"description, a verbose description of the workspace"`
}

func (ws Workspace) GetName() string {
	return ws.Name
}

func (ws Workspace) Map(ctx context.Context) map[string]interface{} {
	return mapper.MarshalMapWithContext(ctx, ws)
}

func (ws Workspace) Dfl(ctx context.Context) string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ws.Map(ctx) {
		dict[dfl.Literal{Value: k}] = dfl.Literal{Value: v}
	}
	return dfl.Dictionary{Nodes: dict}.Dfl(dfl.DefaultQuotes, false, 0)
}

var WorkspaceType = reflect.TypeOf(Workspace{})

func NewDefaultWorkspace() *Workspace {
	return &Workspace{
		Name:        "default",
		Title:       "default",
		Description: "Default Workspace",
	}
}
