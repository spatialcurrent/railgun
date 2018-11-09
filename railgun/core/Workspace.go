// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package core

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"reflect"
)

type Workspace struct {
	Name        string `rest:"name, the unique name of the workspace" required:"yes"`
	Title       string `rest:"title, the title of the workspace"`
	Description string `rest:"description, a verbose description of the workspace"`
}

func (ws Workspace) GetName() string {
	return ws.Name
}

func (ws Workspace) Map() map[string]interface{} {
	return map[string]interface{}{
		"name":        ws.Name,
		"title":       ws.Title,
		"description": ws.Description,
	}
}

func (ws Workspace) Dfl() string {
	dict := map[dfl.Node]dfl.Node{}
	for k, v := range ws.Map() {
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
