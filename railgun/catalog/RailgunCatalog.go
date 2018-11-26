// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package catalog

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun/cache"
	"github.com/spatialcurrent/railgun/railgun/core"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/parser"
	"github.com/spatialcurrent/railgun/railgun/util"
	"github.com/spatialcurrent/viper"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

type RailgunCatalog struct {
	*Catalog
}

func NewRailgunCatalog() *RailgunCatalog {

	catalog := &RailgunCatalog{
		Catalog: &Catalog{
			mutex:   &sync.Mutex{},
			objects: map[string]interface{}{},
			indices: map[string]map[string]int{},
		},
	}

	return catalog
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
	_, uriPath := grw.SplitUri(uriSuffix)
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
	ds := &core.DataStore{
		Workspace:   workspace,
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		Uri:         uriNode,
		Format:      format,
		Compression: compression,
		Extent:      extent,
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
			return &core.Layer{}, errors.Wrap(err, "error parsing process expression")
		}
		lyr.Node = node
	}
	return lyr, nil
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
	s := &core.Service{
		Name:        name,
		Title:       coalesce(title, name),
		Description: coalesce(description, title, name),
		DataStore:   datastore,
		Process:     process,
		Defaults:    defaults,
		Tags:        tags,
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

func (c *RailgunCatalog) LoadFromUri(uri string, logWriter grw.ByteWriteCloser, errorWriter grw.ByteWriteCloser, s3_client *s3.S3) error {

	logWriter.WriteLine(fmt.Sprintf("* loading catalog from %s", uri))

	raw, err := func(uri string) (interface{}, error) {

		_, uriPath := grw.SplitUri(uri)

		name, format, compression := util.SplitNameFormatCompression(filepath.Base(uriPath))
		if len(name) == 0 {
			return nil, &rerrors.ErrInvalidParameter{Name: "uri", Value: uri}
		}
		if len(format) == 0 {
			return nil, &rerrors.ErrInvalidConfig{Name: "uri", Value: uri}
		}

		reader, _, err := grw.ReadFromResource(uri, compression, 4096, false, s3_client)
		if err != nil {
			return nil, err
		}

		inputBytes, err := reader.ReadAllAndClose()
		if err != nil {
			return nil, err
		}

		if len(inputBytes) == 0 {
			return nil, nil
		}

		logWriter.WriteLine(fmt.Sprintf("* catalog is %d bytes", len(inputBytes)))

		inputType, err := gss.GetType(inputBytes, format)
		if err != nil {
			return nil, err
		}

		inputObject, err := gss.DeserializeBytes(inputBytes, format, gss.NoHeader, "", false, gss.NoSkip, gss.NoLimit, inputType, false)
		if err != nil {
			return nil, err
		}

		return inputObject, nil

	}(uri)

	if err != nil {
		return errors.Wrap(err, "error loading catalog")
	}

	if raw == nil {
		logWriter.WriteLine(fmt.Sprint("* catalog was empty"))
		return nil
	}

	if t := reflect.TypeOf(raw); t.Kind() == reflect.Map {
		v := reflect.ValueOf(raw)

		logWriter.WriteLine(fmt.Sprintf("* catalog has keys %v", v.MapKeys()))

		key := "Workspace"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseWorkspace(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workspace"))
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded workspace with name " + obj.Name)
				}
			}
		}

		key = "DataStore"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseDataStore(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading data store"))
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded data store with name " + obj.Name)
				}
			}
		}

		key = "Layer"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseLayer(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading layer"))
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded layer with name " + obj.Name)
				}
			}
		}

		key = "Process"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseProcess(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading process"))
						errorWriter.WriteError(err)
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded process with name " + obj.Name)
				}
			}
		}

		key = "Service"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseService(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading service"))
						errorWriter.WriteError(err)
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded service with name " + obj.Name)
				}
			}
		}

		key = "Job"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseJob(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading job"))
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded job with name " + obj.Name)
				}
			}
		}

		key = "Workflow"
		if list := v.MapIndex(reflect.ValueOf(key)); list.IsValid() {
			listValue := reflect.ValueOf(list.Interface())
			listType := listValue.Type()
			if listType.Kind() == reflect.Array || listType.Kind() == reflect.Slice {
				length := listValue.Len()
				for i := 0; i < length; i++ {
					m := listValue.Index(i).Interface()
					obj, err := c.ParseWorkflow(m)
					if err != nil {
						errorWriter.WriteError(errors.Wrap(&rerrors.ErrInvalidObject{Value: m}, "error loading workflow"))
						errorWriter.WriteError(err)
						continue
					}
					c.Add(obj)
					logWriter.WriteLine("* loaded workflow with name " + obj.Name)
				}
			}
		}

	}

	return nil
}

func (c *RailgunCatalog) LoadFromViper(v *viper.Viper) error {

	workspacesByName, err := func(workspaces []string) (map[string]*core.Workspace, error) {
		workspacesByName := map[string]*core.Workspace{}
		for _, str := range workspaces {
			if !strings.HasPrefix(str, "{") {
				return workspacesByName, &rerrors.ErrInvalidConfig{Name: "workspace", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return workspacesByName, &rerrors.ErrInvalidConfig{Name: "workspace", Value: str}
			}
			ws, err := c.ParseWorkspace(m)
			if err != nil {
				return workspacesByName, errors.Wrap(err, "error parsing workspace")
			}
			workspacesByName[ws.Name] = ws
		}
		if _, ok := workspacesByName["default"]; !ok {
			workspacesByName["default"] = core.NewDefaultWorkspace()
		}
		return workspacesByName, nil
	}(v.GetStringArray("workspace"))

	if err != nil {
		return err
	}

	for _, workspace := range workspacesByName {
		c.Add(workspace)
	}

	datastoresByName, err := func(datastores []string, configSkipErrors bool) (map[string]*core.DataStore, error) {
		datastoresByName := map[string]*core.DataStore{}
		for _, str := range datastores {
			if !strings.HasPrefix(str, "{") {
				_, uriPath := grw.SplitUri(str)
				name, format, compression := util.SplitNameFormatCompression(filepath.Base(uriPath))
				if len(name) == 0 {
					return datastoresByName, &rerrors.ErrInvalidConfig{Name: "datastore", Value: str}
				}
				if len(format) == 0 {
					return datastoresByName, &rerrors.ErrInvalidConfig{Name: "datastore", Value: str}
				}
				ds := &core.DataStore{
					Name:        name,
					Title:       name,
					Description: name,
					Uri:         &dfl.Literal{Value: str},
					Format:      format,
					Compression: compression,
					Extent:      make([]float64, 0),
				}
				datastoresByName[ds.Name] = ds
			} else {
				_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
				if err != nil {
					return datastoresByName, &rerrors.ErrInvalidConfig{Name: "datastore", Value: str}
				}
				ds, err := c.ParseDataStore(m)
				if err != nil {
					if configSkipErrors {
						fmt.Fprint(os.Stderr, err.Error()+"\n")
						continue
					} else {
						return datastoresByName, errors.Wrap(err, "error parsing data store")
					}
				}
				datastoresByName[ds.Name] = ds
			}
		}
		return datastoresByName, nil
	}(v.GetStringArray("datastore"), v.GetBool("config-skip-errors"))

	if err != nil {
		return err
	}

	for _, datastore := range datastoresByName {
		c.Add(datastore)
	}

	layersByName, err := func(layers []string) (map[string]*core.Layer, error) {
		layersByName := map[string]*core.Layer{}
		for _, str := range layers {
			if !strings.HasPrefix(str, "{") {
				return layersByName, &rerrors.ErrInvalidConfig{Name: "layer", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return layersByName, &rerrors.ErrInvalidConfig{Name: "layer", Value: str}
			}
			l, err := c.ParseLayer(m)
			if err != nil {
				return layersByName, errors.Wrap(err, "error parsing layer")
			}
			layersByName[l.Name] = l
		}
		return layersByName, nil
	}(v.GetStringArray("layer"))

	if err != nil {
		return err
	}

	for _, layer := range layersByName {
		c.Add(layer)
	}

	processesByName, err := func(processes []string) (map[string]*core.Process, error) {
		processesByName := map[string]*core.Process{}
		for _, str := range processes {
			if !strings.HasPrefix(str, "{") {
				return processesByName, &rerrors.ErrInvalidConfig{Name: "process", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return processesByName, &rerrors.ErrInvalidConfig{Name: "process", Value: str}
			}
			p, err := c.ParseProcess(m)
			if err != nil {
				return processesByName, errors.Wrap(err, "error parsing process")
			}
			processesByName[p.Name] = p
		}
		return processesByName, nil
	}(v.GetStringArray("process"))

	if err != nil {
		return err
	}

	for _, process := range processesByName {
		c.Add(process)
	}

	servicesByName, err := func(services []string, configSkipErrors bool) (map[string]*core.Service, error) {
		servicesByName := map[string]*core.Service{}
		for _, str := range services {
			if !strings.HasPrefix(str, "{") {
				return servicesByName, &rerrors.ErrInvalidConfig{Name: "service", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return servicesByName, &rerrors.ErrInvalidConfig{Name: "service", Value: str}
			}
			s, err := c.ParseService(m)
			if err != nil {
				if configSkipErrors {
					fmt.Fprint(os.Stderr, err.Error()+"\n")
					continue
				} else {
					return servicesByName, errors.Wrap(err, "error parsing service")
				}
			}
			servicesByName[s.Name] = s
		}
		return servicesByName, nil
	}(v.GetStringArray("service"), v.GetBool("config-skip-errors"))

	if err != nil {
		return err
	}

	for _, service := range servicesByName {
		c.Add(service)
	}

	jobsByName, err := func(jobs []string, configSkipErrors bool) (map[string]*core.Job, error) {
		jobsByName := map[string]*core.Job{}
		for _, str := range jobs {
			if !strings.HasPrefix(str, "{") {
				return jobsByName, &rerrors.ErrInvalidConfig{Name: "job", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return jobsByName, &rerrors.ErrInvalidConfig{Name: "job", Value: str}
			}
			j, err := c.ParseJob(m)
			if err != nil {
				if configSkipErrors {
					fmt.Fprint(os.Stderr, err.Error()+"\n")
					continue
				} else {
					return jobsByName, errors.Wrap(err, "error parsing job")
				}
			}
			jobsByName[j.Name] = j
		}
		return jobsByName, nil
	}(v.GetStringArray("job"), v.GetBool("config-skip-errors"))

	if err != nil {
		return err
	}

	for _, job := range jobsByName {
		c.Add(job)
	}

	workflowsByName, err := func(workflows []string, configSkipErrors bool) (map[string]*core.Workflow, error) {
		workflowsByName := map[string]*core.Workflow{}
		for _, str := range workflows {
			if !strings.HasPrefix(str, "{") {
				return workflowsByName, &rerrors.ErrInvalidConfig{Name: "workflow", Value: str}
			}
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				return workflowsByName, &rerrors.ErrInvalidConfig{Name: "workflow", Value: str}
			}
			wf, err := c.ParseWorkflow(m)
			if err != nil {
				if configSkipErrors {
					fmt.Fprint(os.Stderr, err.Error()+"\n")
					continue
				} else {
					return workflowsByName, errors.Wrap(err, "error parsing workflow")
				}
			}
			workflowsByName[wf.Name] = wf
		}
		return workflowsByName, nil
	}(v.GetStringArray("workflow"), v.GetBool("config-skip-errors"))

	if err != nil {
		return err
	}

	for _, workflow := range workflowsByName {
		c.Add(workflow)
	}

	return nil

}

func (c *RailgunCatalog) SaveToUri(uri string, s3_client *s3.S3) error {

	data := c.Dump()

	err := func(data map[string]interface{}, uri string) error {

		scheme, uriPath := grw.SplitUri(uri)

		name, format, compression := util.SplitNameFormatCompression(filepath.Base(uriPath))
		if len(name) == 0 {
			return &rerrors.ErrInvalidConfig{Name: "catalog-uri", Value: uri}
		}
		if len(format) == 0 {
			return &rerrors.ErrInvalidConfig{Name: "catalog-uri", Value: uri}
		}

		b, err := gss.SerializeBytes(data, format, gss.NoHeader, gss.NoLimit)
		if err != nil {
			return errors.Wrap(err, "error serializing catalog")
		}

		if scheme == "s3" {
			i := strings.Index(uriPath, "/")
			if i == -1 {
				return errors.New("s3 path missing bucket")
			}
			err := grw.UploadS3Object(uriPath[0:i], uriPath[i+1:], bytes.NewBuffer(b), s3_client)
			if err != nil {
				return errors.Wrap(err, "error uploading new version of catalog to S3")
			}
			return nil
		}

		outputWriter, err := grw.WriteToResource(uri, compression, false, nil)
		if err != nil {
			return errors.Wrap(err, "error opening writer")
		}

		_, err = outputWriter.Write(b)
		if err != nil {
			outputWriter.Close()
			return errors.Wrap(err, "error saving catalog")
		}

		err = outputWriter.Close()
		if err != nil {
			return errors.Wrap(err, "error closing file")
		}

		return nil

	}(data, uri)

	if err != nil {
		return errors.Wrap(err, "error saving catalog")
	}

	return nil
}
