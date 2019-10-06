// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/spatialcurrent/viper"
)

import (
	"github.com/spatialcurrent/go-dfl/pkg/dfl"
	"github.com/spatialcurrent/go-reader-writer/pkg/grw"
	"github.com/spatialcurrent/go-reader-writer/pkg/splitter"
	"github.com/spatialcurrent/go-simple-serializer/pkg/gss"
	"github.com/spatialcurrent/go-sync-logger/pkg/gsl"
	"github.com/spatialcurrent/go-try-get/pkg/gtg"
)

import (
	"github.com/spatialcurrent/railgun/pkg/cache"
	"github.com/spatialcurrent/railgun/pkg/core"
	rerrors "github.com/spatialcurrent/railgun/pkg/errors"
	"github.com/spatialcurrent/railgun/pkg/parser"
	"github.com/spatialcurrent/railgun/pkg/util"
)

type RailgunCatalog struct {
	*Catalog
}

func NewRailgunCatalog() *RailgunCatalog {
	return &RailgunCatalog{Catalog: NewCatalog()}
}

func (c *RailgunCatalog) ParseWorkspace(obj interface{}) (*core.Workspace, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return &core.Workspace{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	ws := &core.Workspace{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
	}
	return ws, nil
}

func (c *RailgunCatalog) ParseDataStore(obj interface{}) (*core.DataStore, error) {
	uri := gtg.TryGetString(obj, "uri", "")
	if len(uri) == 0 {
		return &core.DataStore{}, &rerrors.ErrMissingRequiredParameter{Name: "uri"}
	}
	uriNode, err := dfl.ParseCompile(uri)
	if err != nil {
		return &core.DataStore{}, err
	}
	uriSuffix := ""
	switch n := uriNode.(type) {
	case dfl.Literal:
		switch v := n.Value.(type) {
		case string:
			uriSuffix = v
		}
	case dfl.Concat:
		uriSuffix = n.Suffix()
	}
	_, uriPath := splitter.SplitUri(uriSuffix)
	uriName, uriFormat, uriCompression := util.SplitNameFormatCompression(filepath.Base(uriPath))
	name := gtg.TryGetString(obj, "name", uriName)
	if len(name) == 0 {
		return &core.DataStore{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	format := gtg.TryGetString(obj, "format", uriFormat)
	if len(format) == 0 {
		return &core.DataStore{}, &rerrors.ErrMissingRequiredParameter{Name: "format"}
	}
	compression := gtg.TryGetString(obj, "compression", uriCompression)
	extent, err := parser.ParseFloat64Array(obj, "extent")
	if err != nil {
		return &core.DataStore{}, err
	}
	workspaceName := gtg.TryGetString(obj, "workspace", "")
	if len(workspaceName) == 0 {
		return &core.DataStore{}, &rerrors.ErrMissingRequiredParameter{Name: "workspace"}
	}
	workspace, found := c.GetWorkspace(workspaceName)
	if !found {
		return &core.DataStore{}, &rerrors.ErrMissingObject{Type: "workspace", Name: workspaceName}
	}
	vars, err := parser.ParseMap(obj, "vars")
	if err != nil {
		return &core.DataStore{}, err
	}
	var filterNode dfl.Node
	if filterString := gtg.TryGetString(obj, "filter", ""); len(filterString) > 0 {
		n, err := dfl.ParseCompile(filterString)
		if err != nil {
			return &core.DataStore{}, errors.Wrap(err, "error parsing data store filter")
		}
		filterNode = n
	}
	ds := &core.DataStore{
		Workspace:   workspace,
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Uri:         uriNode,
		Format:      format,
		Compression: compression,
		Extent:      extent,
		Vars:        vars,
		Filter:      filterNode,
	}
	return ds, nil
}

func (c *RailgunCatalog) ParseLayer(obj interface{}) (*core.Layer, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return &core.Layer{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	datastoreName := gtg.TryGetString(obj, "datastore", "")
	if len(datastoreName) == 0 {
		return &core.Layer{}, &rerrors.ErrMissingRequiredParameter{Name: "datastore"}
	}
	datastore, found := c.GetDataStore(datastoreName)
	if !found {
		return &core.Layer{}, &rerrors.ErrMissingObject{Type: "datastore", Name: datastoreName}
	}
	extent, err := parser.ParseFloat64Array(obj, "extent")
	if err != nil {
		return &core.Layer{}, err
	}
	defaults, err := parser.ParseMap(obj, "defaults")
	if err != nil {
		return &core.Layer{}, err
	}
	tags, err := parser.ParseStringArray(obj, "tags")
	if err != nil {
		return &core.Layer{}, err
	}
	lyr := &core.Layer{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		DataStore:   datastore,
		Node:        nil,
		Defaults:    defaults,
		Extent:      extent,
		Tags:        tags,
		Cache:       cache.NewCache(),
	}
	expression := gtg.TryGetString(obj, "expression", "")
	if len(expression) > 0 {
		node, err := dfl.ParseCompile(expression)
		if err != nil {
			return &core.Layer{}, errors.Wrap(err, "error parsing layer expression")
		}
		lyr.Node = node
	}
	return lyr, nil
}

func (c *RailgunCatalog) ParseFunction(obj interface{}) (*core.Function, error) {
	expression := gtg.TryGetString(obj, "expression", "")
	if len(expression) == 0 {
		return &core.Function{}, &rerrors.ErrMissingRequiredParameter{Name: "expression"}
	}
	node, err := dfl.ParseCompile(expression)
	if err != nil {
		return &core.Function{}, errors.Wrap(err, "error parsing function expression")
	}
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return &core.Function{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	aliases, err := parser.ParseStringArray(obj, "aliases")
	if err != nil {
		return &core.Function{}, err
	}
	tags, err := parser.ParseStringArray(obj, "tags")
	if err != nil {
		return &core.Function{}, err
	}
	f := &core.Function{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Aliases:     aliases,
		Node:        node,
		Tags:        tags,
	}
	return f, nil
}

func (c *RailgunCatalog) ParseProcess(obj interface{}) (*core.Process, error) {
	expression := gtg.TryGetString(obj, "expression", "")
	if len(expression) == 0 {
		return &core.Process{}, &rerrors.ErrMissingRequiredParameter{Name: "expression"}
	}
	node, err := dfl.ParseCompile(expression)
	if err != nil {
		return &core.Process{}, errors.Wrap(err, "error parsing process expression")
	}
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return &core.Process{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	tags, err := parser.ParseStringArray(obj, "tags")
	if err != nil {
		return &core.Process{}, err
	}
	p := &core.Process{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Node:        node,
		Tags:        tags,
	}
	return p, nil
}

func (c *RailgunCatalog) ParseService(obj interface{}) (*core.Service, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return &core.Service{}, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	datastoreName := gtg.TryGetString(obj, "datastore", "")
	if len(datastoreName) == 0 {
		return &core.Service{}, &rerrors.ErrMissingRequiredParameter{Name: "datastore"}
	}
	datastore, found := c.GetDataStore(datastoreName)
	if !found {
		return &core.Service{}, &rerrors.ErrMissingObject{Type: "datastore", Name: datastoreName}
	}
	processName := gtg.TryGetString(obj, "process", "")
	if len(processName) == 0 {
		return &core.Service{}, &rerrors.ErrMissingRequiredParameter{Name: "process"}
	}
	process, found := c.GetProcess(processName)
	if !found {
		return &core.Service{}, &rerrors.ErrMissingObject{Type: "process", Name: processName}
	}
	defaults, err := parser.ParseMap(obj, "defaults")
	if err != nil {
		return &core.Service{}, err
	}
	tags, err := parser.ParseStringArray(obj, "tags")
	if err != nil {
		return &core.Service{}, err
	}
	var transformNode dfl.Node
	if transformString := gtg.TryGetString(obj, "transform", ""); len(transformString) > 0 {
		n, err := dfl.ParseCompile(transformString)
		if err != nil {
			return &core.Service{}, errors.Wrap(err, "error parsing service transform")
		}
		transformNode = n
	}
	s := &core.Service{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		DataStore:   datastore,
		Process:     process,
		Defaults:    defaults,
		Tags:        tags,
		Transform:   transformNode,
	}
	return s, nil
}

func (c *RailgunCatalog) ParseJob(obj interface{}) (*core.Job, error) {
	name := gtg.TryGetString(obj, "name", "")
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	serviceName := gtg.TryGetString(obj, "service", "")
	if len(serviceName) == 0 {
		return &core.Job{}, &rerrors.ErrMissingRequiredParameter{Name: "service"}
	}
	service, found := c.GetService(serviceName)
	if !found {
		return &core.Job{}, &rerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}
	variables, err := parser.ParseMap(obj, "variables")
	if err != nil {
		return &core.Job{}, errors.Wrap(err, (&rerrors.ErrInvalidConfig{Name: "job", Value: obj}).Error())
	}
	j := &core.Job{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Service:     service,
		Variables:   variables,
	}
	return j, nil
}

func (c *RailgunCatalog) ParseWorkflow(obj interface{}) (*core.Workflow, error) {
	name := gtg.TryGetString(obj, "name", "")
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	jobNames, err := parser.ParseStringArray(obj, "jobs")
	if err != nil {
		return &core.Workflow{}, err
	}
	if len(jobNames) == 0 {
		return &core.Workflow{}, &rerrors.ErrMissingRequiredParameter{Name: "jobs"}
	}
	jobs, err := c.GetJobs(jobNames)
	if err != nil {
		return &core.Workflow{}, err
	}
	variables, err := parser.ParseMap(obj, "variables")
	if err != nil {
		return &core.Workflow{}, err
	}
	wf := &core.Workflow{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Variables:   variables,
		Jobs:        jobs,
	}
	return wf, nil
}

func (c *RailgunCatalog) ParseItem(obj interface{}, t reflect.Type) (core.Base, error) {
	switch t {
	case core.WorkspaceType:
		return c.ParseWorkspace(obj)
	case core.DataStoreType:
		return c.ParseDataStore(obj)
	case core.LayerType:
		return c.ParseLayer(obj)
	case core.FunctionType:
		return c.ParseFunction(obj)
	case core.ProcessType:
		return c.ParseProcess(obj)
	case core.ServiceType:
		return c.ParseService(obj)
	case core.JobType:
		return c.ParseJob(obj)
	case core.WorkflowType:
		return c.ParseWorkflow(obj)
	}
	return nil, &rerrors.ErrInvalidType{Value: t}
}

func (c *RailgunCatalog) GetWorkspace(name string) (*core.Workspace, bool) {
	obj, ok := c.Get(name, core.WorkspaceType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Workspace), ok
}

func (c *RailgunCatalog) GetDataStore(name string) (*core.DataStore, bool) {
	obj, ok := c.Get(name, core.DataStoreType)
	if !ok {
		return nil, false
	}
	return obj.(*core.DataStore), ok
}

func (c *RailgunCatalog) GetLayer(name string) (*core.Layer, bool) {
	obj, ok := c.Get(name, core.LayerType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Layer), ok
}

func (c *RailgunCatalog) GetFunction(name string) (*core.Function, bool) {
	obj, ok := c.Get(name, core.FunctionType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Function), ok
}

func (c *RailgunCatalog) GetProcess(name string) (*core.Process, bool) {
	obj, ok := c.Get(name, core.ProcessType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Process), ok
}

func (c *RailgunCatalog) GetService(name string) (*core.Service, bool) {
	obj, ok := c.Get(name, core.ServiceType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Service), ok
}

func (c *RailgunCatalog) GetJob(name string) (*core.Job, bool) {
	obj, ok := c.Get(name, core.JobType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Job), ok
}

func (c *RailgunCatalog) GetJobs(names []string) ([]*core.Job, error) {
	jobs := make([]*core.Job, 0, len(names))
	for _, name := range names {
		job, ok := c.GetJob(name)
		if !ok {
			return jobs, &rerrors.ErrMissingObject{Type: "job", Name: name}
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (c *RailgunCatalog) GetWorkflow(name string) (*core.Workflow, bool) {
	obj, ok := c.Get(name, core.WorkflowType)
	if !ok {
		return nil, false
	}
	return obj.(*core.Workflow), ok
}

func (c *RailgunCatalog) GetItem(name string, t reflect.Type) (core.Base, bool) {
	switch t {
	case core.WorkspaceType:
		return c.GetWorkspace(name)
	case core.DataStoreType:
		return c.GetDataStore(name)
	case core.LayerType:
		return c.GetLayer(name)
	case core.FunctionType:
		return c.GetFunction(name)
	case core.ProcessType:
		return c.GetProcess(name)
	case core.ServiceType:
		return c.GetService(name)
	case core.JobType:
		return c.GetJob(name)
	case core.WorkflowType:
		return c.GetWorkflow(name)
	}
	return nil, false
}

func (c *RailgunCatalog) DeleteWorkspace(name string) error {
	if _, ok := c.GetWorkspace(name); !ok {
		return &rerrors.ErrMissingObject{Type: "workspace", Name: name}
	}
	for _, ds := range c.ListDataStores() {
		if ds.Workspace != nil && ds.Workspace.Name == name {
			return &rerrors.ErrDependent{DependentType: "data store", DependentName: ds.Name, Type: "workspace", Name: name}
		}
	}
	return c.Catalog.Delete(name, core.WorkspaceType)
}

func (c *RailgunCatalog) DeleteDataStore(name string) error {
	if _, ok := c.GetDataStore(name); !ok {
		return &rerrors.ErrMissingObject{Type: "data store", Name: name}
	}
	for _, l := range c.ListLayers() {
		if l.DataStore != nil && l.DataStore.Name == name {
			return &rerrors.ErrDependent{DependentType: "layer", DependentName: l.Name, Type: "data store", Name: name}
		}
	}
	for _, s := range c.ListServices() {
		if s.DataStore != nil && s.DataStore.Name == name {
			return &rerrors.ErrDependent{DependentType: "service", DependentName: s.Name, Type: "data store", Name: name}
		}
	}
	return c.Delete(name, core.DataStoreType)
}

func (c *RailgunCatalog) DeleteLayer(name string) error {
	if _, ok := c.GetLayer(name); !ok {
		return &rerrors.ErrMissingObject{Type: "layer", Name: name}
	}
	return c.Delete(name, core.LayerType)
}

func (c *RailgunCatalog) DeleteFunction(name string) error {
	if _, ok := c.GetFunction(name); !ok {
		return &rerrors.ErrMissingObject{Type: "function", Name: name}
	}
	return c.Delete(name, core.FunctionType)
}

func (c *RailgunCatalog) DeleteProcess(name string) error {
	if _, ok := c.GetProcess(name); !ok {
		return &rerrors.ErrMissingObject{Type: "process", Name: name}
	}
	for _, s := range c.ListServices() {
		if s.Process.Name == name {
			return &rerrors.ErrDependent{DependentType: "service", DependentName: s.Name, Type: "process", Name: name}
		}
	}
	return c.Delete(name, core.ProcessType)
}

func (c *RailgunCatalog) DeleteService(name string) error {
	if _, ok := c.GetService(name); !ok {
		return &rerrors.ErrMissingObject{Type: "service", Name: name}
	}
	for _, j := range c.ListJobs() {
		if j.Service != nil && j.Service.Name == name {
			return &rerrors.ErrDependent{DependentType: "job", DependentName: j.Name, Type: "service", Name: name}
		}
	}
	return c.Delete(name, core.ServiceType)
}

func (c *RailgunCatalog) DeleteJob(name string) error {
	if _, ok := c.GetJob(name); !ok {
		return &rerrors.ErrMissingObject{Type: "job", Name: name}
	}
	for _, workflow := range c.ListWorkflows() {
		for _, job := range workflow.Jobs {
			if job.Name == name {
				return &rerrors.ErrDependent{DependentType: "workflow", DependentName: workflow.Name, Type: "job", Name: name}
			}
		}
	}
	return c.Delete(name, core.JobType)
}

func (c *RailgunCatalog) DeleteWorkflow(name string) error {
	return c.Delete(name, core.WorkflowType)
}

func (c *RailgunCatalog) DeleteItem(name string, t reflect.Type) error {
	switch t {
	case core.WorkspaceType:
		return c.DeleteWorkspace(name)
	case core.DataStoreType:
		return c.DeleteDataStore(name)
	case core.LayerType:
		return c.DeleteLayer(name)
	case core.FunctionType:
		return c.DeleteFunction(name)
	case core.ProcessType:
		return c.DeleteProcess(name)
	case core.ServiceType:
		return c.DeleteService(name)
	case core.JobType:
		return c.DeleteJob(name)
	case core.WorkflowType:
		return c.DeleteWorkflow(name)
	}
	return &rerrors.ErrInvalidType{Value: t}
}

func (c *RailgunCatalog) ListWorkspaces() []*core.Workspace {
	return c.Catalog.List(core.WorkspaceType).([]*core.Workspace)
}

func (c *RailgunCatalog) ListDataStores() []*core.DataStore {
	return c.Catalog.List(core.DataStoreType).([]*core.DataStore)
}

func (c *RailgunCatalog) ListLayers() []*core.Layer {
	return c.Catalog.List(core.LayerType).([]*core.Layer)
}

func (c *RailgunCatalog) ListFunctions() []*core.Function {
	return c.Catalog.List(core.FunctionType).([]*core.Function)
}

func (c *RailgunCatalog) ListProcesses() []*core.Process {
	return c.Catalog.List(core.ProcessType).([]*core.Process)
}

func (c *RailgunCatalog) ListServices() []*core.Service {
	return c.Catalog.List(core.ServiceType).([]*core.Service)
}

func (c *RailgunCatalog) ListJobs() []*core.Job {
	return c.Catalog.List(core.JobType).([]*core.Job)
}

func (c *RailgunCatalog) ListWorkflows() []*core.Workflow {
	return c.Catalog.List(core.WorkflowType).([]*core.Workflow)
}
