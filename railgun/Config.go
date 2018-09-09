// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package railgun

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-adaptive-functions/af"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun/railgunerrors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	"reflect"
	"strings"
)

var ConfigNoFile = []string{
	"aws-access-key-id",
	"aws-default-region",
	"aws-profile",
	"aws-region",
	"aws-secret-access-key",
	"aws-security-token",
	"aws-session-token",
	"config-uri"}

type Config struct {
	*viper.Viper
	Funcs            dfl.FunctionMap
	workspacesList   *[]Workspace
	workspacesByName map[string]Workspace
	datastoresList   *[]DataStore
	datastoresByName map[string]DataStore
	layersList       *[]Layer
	layersByName     map[string]Layer
	processesList    *[]Process
	processesByName  map[string]Process
	servicesList     *[]Service
	servicesByName   map[string]Service
	jobsList         *[]Job
	jobsByName       map[string]Job
}

func (c *Config) Coalesce(strs ...string) string {
	for _, str := range strs {
		if len(str) > 0 {
			return str
		}
	}
	return ""
}

func NewConfig(cmd *cobra.Command) *Config {

	v := viper.New()
	v.BindPFlags(cmd.PersistentFlags())
	v.BindPFlags(cmd.Flags())
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv() // set environment variables to overwrite config
	MergeConfigs(v, v.GetStringArray("config-uri"))

	config := &Config{
		Viper: v,
		Funcs: dfl.NewFuntionMapWithDefaults()}

	return config
}

func (c *Config) Reload() error {
	err := c.ReloadWorkspaces()
	if err != nil {
		return errors.Wrap(err, "error reloading workspaces")
	}
	err = c.ReloadDataStores()
	if err != nil {
		return errors.Wrap(err, "error reloading data stores")
	}
	err = c.ReloadLayers()
	if err != nil {
		return errors.Wrap(err, "error reloading layers")
	}
	err = c.ReloadProcesses()
	if err != nil {
		return errors.Wrap(err, "error reloading processes")
	}
	err = c.ReloadServices()
	if err != nil {
		return errors.Wrap(err, "error reloading services")
	}
	err = c.ReloadJobs()
	if err != nil {
		return errors.Wrap(err, "error reloading jobs")
	}
	return nil
}

func (c *Config) ReloadWorkspaces() error {

	workspacesByName := map[string]Workspace{}
	workspaces := c.GetStringArray("workspace")
	for _, str := range workspaces {
		if !strings.HasPrefix(str, "{") {
			return &railgunerrors.ErrInvalidConfig{Name: "workspace", Value: str}
		}
		_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
		if err != nil {
			return &railgunerrors.ErrInvalidConfig{Name: "workspace", Value: str}
		}
		name := gtg.TryGetString(m, "name", "")
		if len(name) == 0 {
			return &railgunerrors.ErrInvalidConfig{Name: "workspace", Value: str}
		}
		title := gtg.TryGetString(m, "title", "")
		description := gtg.TryGetString(m, "description", "")
		ws := Workspace{
			Name:        name,
			Title:       c.Coalesce(title, name),
			Description: c.Coalesce(description, title, name),
		}
		workspacesByName[ws.Name] = ws
	}

	if _, ok := workspacesByName["default"]; !ok {
		workspacesByName["default"] = Workspace{
			Name:        "default",
			Title:       "default",
			Description: "Default Workspace",
		}
	}

	workspacesList := make([]Workspace, 0, len(workspacesByName))
	for _, ws := range workspacesByName {
		workspacesList = append(workspacesList, ws)
	}

	// Save to Object
	c.workspacesByName = workspacesByName
	c.workspacesList = &workspacesList

	return nil
}

func (c *Config) ReloadDataStores() error {

	datastoresList := make([]DataStore, 0)
	datastoresByName := map[string]DataStore{}
	datastores := c.GetStringArray("datastore")
	for _, str := range datastores {
		if !strings.HasPrefix(str, "{") {
			_, uriPath := grw.SplitUri(str)
			name, format, compression := SplitNameFormatCompression(filepath.Base(uriPath))
			if len(name) == 0 {
				return &railgunerrors.ErrInvalidConfig{Name: "datastore", Value: str}
			}
			if len(format) == 0 {
				return &railgunerrors.ErrInvalidConfig{Name: "datastore", Value: str}
			}
			ds := DataStore{
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
			_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
			if err != nil {
				return &railgunerrors.ErrInvalidConfig{Name: "datastore", Value: str}
			}
			ds, err := c.ParseDataStore(m)
			if err != nil {
				return errors.Wrap(err, "error parsing data store")
			}
			datastoresList = append(datastoresList, ds)
			datastoresByName[ds.Name] = ds
		}
	}

	// Save to Object
	c.datastoresByName = datastoresByName
	c.datastoresList = &datastoresList

	return nil
}

func (c *Config) ReloadLayers() error {

	layersList := make([]Layer, 0)
	layersByName := map[string]Layer{}
	layers := c.GetStringArray("layer")
	for _, str := range layers {
		if !strings.HasPrefix(str, "{") {
			return &railgunerrors.ErrInvalidConfig{Name: "layer", Value: str}
		}
		_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
		if err != nil {
			return &railgunerrors.ErrInvalidConfig{Name: "layer", Value: str}
		}
		l, err := c.ParseLayer(m)
		if err != nil {
			return errors.Wrap(err, "error parsing layer")
		}
		layersList = append(layersList, l)
		layersByName[l.Name] = l
	}

	// Save to Object
	c.layersByName = layersByName
	c.layersList = &layersList

	return nil
}

func (c *Config) ReloadProcesses() error {

	processesList := make([]Process, 0)
	processesByName := map[string]Process{}
	processes := c.GetStringArray("process")
	for _, str := range processes {
		if !strings.HasPrefix(str, "{") {
			return &railgunerrors.ErrInvalidConfig{Name: "process", Value: str}
		}

		_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
		if err != nil {
			return &railgunerrors.ErrInvalidConfig{Name: "process", Value: str}
		}
		p, err := c.ParseProcess(m)
		if err != nil {
			return errors.Wrap(err, "error parsing process")
		}
		processesList = append(processesList, p)
		processesByName[p.Name] = p
	}

	// Save to Object
	c.processesByName = processesByName
	c.processesList = &processesList

	return nil

}

func (c *Config) ReloadServices() error {

	servicesList := make([]Service, 0)
	servicesByName := map[string]Service{}
	services := c.GetStringArray("service")
	for _, str := range services {
		if !strings.HasPrefix(str, "{") {
			return &railgunerrors.ErrInvalidConfig{Name: "service", Value: str}
		}
		_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
		if err != nil {
			return &railgunerrors.ErrInvalidConfig{Name: "service", Value: str}
		}
		s, err := c.ParseService(m)
		if err != nil {
			return errors.Wrap(err, "error parsing service")
		}
		servicesList = append(servicesList, s)
		servicesByName[s.Name] = s
	}

	// Save to Object
	c.servicesByName = servicesByName
	c.servicesList = &servicesList

	return nil
}

func (c *Config) ReloadJobs() error {

	jobsList := make([]Job, 0)
	jobsByName := map[string]Job{}
	jobs := c.GetStringArray("job")
	for _, str := range jobs {
		if !strings.HasPrefix(str, "{") {
			return &railgunerrors.ErrInvalidConfig{Name: "job", Value: str}
		}
		_, m, err := dfl.ParseCompileEvaluateMap(str, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
		if err != nil {
			return &railgunerrors.ErrInvalidConfig{Name: "job", Value: str}
		}
		j, err := c.ParseJob(m)
		if err != nil {
			return errors.Wrap(err, "error parsing job")
		}
		jobsList = append(jobsList, j)
		jobsByName[j.Name] = j
	}

	// Save to Object
	c.jobsByName = jobsByName
	c.jobsList = &jobsList

	return nil
}

func (c *Config) Save() error {

	allSettings := c.Viper.AllSettings()
	for _, k := range ConfigNoFile {
		delete(allSettings, k)
	}

	workspaces := make([]string, 0)
	for _, ws := range c.ListWorkspaces() {
		workspaces = append(workspaces, ws.Dfl())
	}
	allSettings["workspace"] = workspaces

	datastores := make([]string, 0)
	for _, ds := range c.ListDataStores() {
		datastores = append(datastores, ds.Dfl())
	}
	allSettings["datastore"] = datastores

	layers := make([]string, 0)
	for _, s := range c.ListLayers() {
		layers = append(layers, s.Dfl())
	}
	allSettings["layer"] = layers

	processes := make([]string, 0)
	for _, p := range c.ListProcesses() {
		processes = append(processes, p.Dfl())
	}
	allSettings["process"] = processes

	services := make([]string, 0)
	for _, s := range c.ListServices() {
		services = append(services, s.Dfl())
	}
	allSettings["service"] = services

	uris := c.GetStringArray("config-uri")
	if len(uris) == 0 {
		return errors.Wrap(&railgunerrors.ErrInvalidConfig{Name: "config-uri", Value: uris}, "error saving config")
	}

	uri := uris[0]

	_, uriPath := grw.SplitUri(uri)

	name, format, compression := SplitNameFormatCompression(filepath.Base(uriPath))
	if len(name) == 0 {
		return errors.Wrap(&railgunerrors.ErrInvalidConfig{Name: "config-uri", Value: uris}, "error saving config")
	}
	if len(format) == 0 {
		return errors.Wrap(&railgunerrors.ErrInvalidConfig{Name: "config-uri", Value: uris}, "error saving config")
	}

	configWriter, err := grw.WriteToResource(uri, compression, false, nil)
	if err != nil {
		return errors.Wrap(err, "error saving config")
	}

	b, err := gss.SerializeBytes(allSettings, format, []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error serializing config")
	}
	_, err = configWriter.Write(b)
	if err != nil {
		configWriter.Close()
		return errors.Wrap(err, "error saving config")
	}
	err = configWriter.Close()
	if err != nil {
		return errors.Wrap(err, "error closing file")
	}
	return nil
}

func (c *Config) Print() error {
	str, err := gss.SerializeString(c.AllSettings(), "properties", []string{}, gss.NoLimit)
	if err != nil {
		return errors.Wrap(err, "error getting all settings for config")
	}
	fmt.Println("=================================================")
	fmt.Println("Configuration:")
	fmt.Println("-------------------------------------------------")
	fmt.Println(str)
	fmt.Println("=================================================")
	return nil

}

func (c *Config) ParseWorkspace(obj interface{}) (Workspace, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return Workspace{}, &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	ws := Workspace{
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
	}
	return ws, nil
}

func (c *Config) AddWorkspace(ws Workspace) error {
	if _, ok := c.workspacesByName[ws.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "workspace", Value: ws.Name}
	}
	c.workspacesByName[ws.Name] = ws
	*c.workspacesList = append(*c.workspacesList, ws)
	return nil
}

func (c *Config) GetWorkspace(name string) (Workspace, bool) {
	ds, ok := c.workspacesByName[name]
	return ds, ok
}

func (c *Config) DeleteWorkspace(name string) error {
	if _, ok := c.workspacesByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "workspace", Name: name}
	}
	for _, ds := range c.ListDataStores() {
		if ds.Workspace.Name == name {
			return &railgunerrors.ErrDependent{DependentType: "data store", DependentName: ds.Name, Type: "workspace", Name: name}
		}
	}
	delete(c.workspacesByName, name)
	existing := *c.workspacesList
	keep := make([]Workspace, len(existing)-1)
	for _, ws := range *c.workspacesList {
		if ws.Name != name {
			keep = append(keep, ws)
		}
	}
	c.workspacesList = &keep
	return nil
}

func (c *Config) ListWorkspaces() []Workspace {
	return *c.workspacesList
}

func (c *Config) ParseMap(obj interface{}, name string) (map[string]interface{}, error) {
	expression := gtg.TryGetString(obj, name, "")
	if len(expression) == 0 {
		return make(map[string]interface{}, 0), nil
	}
	_, m, err := dfl.ParseCompileEvaluateMap(expression, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
	if err != nil {
		return make(map[string]interface{}, 0), err
	}
	if reflect.TypeOf(m).Kind() == reflect.Map {
		if reflect.ValueOf(m).Len() == 0 {
			return map[string]interface{}{}, nil
		}
	}
	return gss.StringifyMapKeys(m).(map[string]interface{}), err
}

func (c *Config) ParseFloat64Array(obj interface{}, name string) ([]float64, error) {
	expression := gtg.TryGetString(obj, name, "")
	if len(expression) == 0 {
		return make([]float64, 0), nil
	}
	_, arr, err := dfl.ParseCompileEvaluate(expression, dfl.NoVars, dfl.NoContext, c.Funcs, dfl.DefaultQuotes)
	if err != nil {
		return make([]float64, 0), err
	}
	extent, err := af.ToFloat64Array.ValidateRun([]interface{}{arr})
	if err != nil {
		return make([]float64, 0), err
	}
	return extent.([]float64), nil
}

func (c *Config) ParseDataStore(obj interface{}) (DataStore, error) {
	uri := gtg.TryGetString(obj, "uri", "")
	if len(uri) == 0 {
		return DataStore{}, &railgunerrors.ErrMissingRequiredParameter{Name: "uri"}
	}
	uriNode, err := dfl.ParseCompile(uri)
	if err != nil {
		return DataStore{}, errors.Wrap(err, (&railgunerrors.ErrInvalidConfig{Name: "datastore", Value: obj}).Error())
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
	uriName, uriFormat, uriCompression := SplitNameFormatCompression(filepath.Base(uriPath))
	name := gtg.TryGetString(obj, "name", uriName)
	if len(name) == 0 {
		return DataStore{}, &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	format := gtg.TryGetString(obj, "format", uriFormat)
	if len(format) == 0 {
		return DataStore{}, errors.Wrap(errors.New("format is blank"), (&railgunerrors.ErrInvalidConfig{Name: "datastore", Value: obj}).Error())
	}
	compression := gtg.TryGetString(obj, "compression", uriCompression)
	extent, err := c.ParseFloat64Array(obj, "extent")
	if err != nil {
		return DataStore{}, errors.Wrap(err, (&railgunerrors.ErrInvalidConfig{Name: "datastore", Value: obj}).Error())
	}
	workspace, found := c.GetWorkspace(gtg.TryGetString(obj, "workspace", ""))
	if !found {
		return DataStore{}, errors.Wrap(errors.New("could not find workspace with name "+gtg.TryGetString(obj, "workspace", "")), (&railgunerrors.ErrInvalidConfig{Name: "datastore", Value: obj}).Error())
	}
	ds := DataStore{
		Workspace:   workspace,
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
		Uri:         uriNode,
		Format:      format,
		Compression: compression,
		Extent:      extent,
	}
	return ds, nil
}

func (c *Config) AddDataStore(ds DataStore) error {
	if _, ok := c.datastoresByName[ds.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "datastore", Value: ds.Name}
	}
	c.datastoresByName[ds.Name] = ds
	*c.datastoresList = append(*c.datastoresList, ds)
	return nil
}

func (c *Config) GetDataStore(name string) (DataStore, bool) {
	ds, ok := c.datastoresByName[name]
	return ds, ok
}

func (c *Config) DeleteDataStore(name string) error {
	if _, ok := c.datastoresByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "data store", Name: name}
	}
	for _, l := range c.ListLayers() {
		if l.DataStore.Name == name {
			return &railgunerrors.ErrDependent{DependentType: "layer", DependentName: l.Name, Type: "data store", Name: name}
		}
	}
	for _, s := range c.ListServices() {
		if s.DataStore.Name == name {
			return &railgunerrors.ErrDependent{DependentType: "service", DependentName: s.Name, Type: "data store", Name: name}
		}
	}
	delete(c.datastoresByName, name)
	existing := *c.datastoresList
	keep := make([]DataStore, len(existing)-1)
	for _, ds := range existing {
		if ds.Name != name {
			keep = append(keep, ds)
		}
	}
	c.datastoresList = &keep
	return nil
}

func (c *Config) ListDataStores() []DataStore {
	return *c.datastoresList
}

func (c *Config) ParseLayer(obj interface{}) (Layer, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return Layer{}, &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	datastoreName := gtg.TryGetString(obj, "datastore", "")
	datastore, found := c.GetDataStore(datastoreName)
	if !found {
		return Layer{}, &railgunerrors.ErrMissingObject{Type: "datastore", Name: datastoreName}
	}
	extent, err := c.ParseFloat64Array(obj, "extent")
	if err != nil {
		return Layer{}, errors.Wrap(err, (&railgunerrors.ErrInvalidConfig{Name: "layer", Value: obj}).Error())
	}
	lyr := Layer{
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
		DataStore:   datastore,
		Extent:      extent,
		Cache:       NewCache(),
	}
	return lyr, nil
}

func (c *Config) AddLayer(l Layer) error {
	if _, ok := c.layersByName[l.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "layer", Value: l.Name}
	}
	c.layersByName[l.Name] = l
	*c.layersList = append(*c.layersList, l)
	return nil
}

func (c *Config) GetLayer(name string) (Layer, bool) {
	p, ok := c.layersByName[name]
	return p, ok
}

func (c *Config) DeleteLayer(name string) error {
	if _, ok := c.layersByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "layer", Name: name}
	}
	delete(c.layersByName, name)
	existing := *c.layersList
	keep := make([]Layer, len(existing)-1)
	for _, lyr := range existing {
		if lyr.Name != name {
			keep = append(keep, lyr)
		}
	}
	c.layersList = &keep
	return nil
}

func (c *Config) ListLayers() []Layer {
	return *c.layersList
}

func (c *Config) ParseProcess(obj interface{}) (Process, error) {
	expression := gtg.TryGetString(obj, "expression", "")
	if len(expression) == 0 {
		return Process{}, errors.Wrap(errors.New("expression is blank"), (&railgunerrors.ErrInvalidConfig{Name: "process", Value: obj}).Error())
	}
	node, err := dfl.ParseCompile(expression)
	if err != nil {
		return Process{}, errors.Wrap(err, "error parsing process expression")
	}
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return Process{}, errors.Wrap(errors.New("name is blank"), (&railgunerrors.ErrInvalidConfig{Name: "datastore", Value: obj}).Error())
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	p := Process{
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
		Node:        node,
	}
	return p, nil
}

func (c *Config) AddProcess(p Process) error {
	if _, ok := c.processesByName[p.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "process", Value: p.Name}
	}
	c.processesByName[p.Name] = p
	*c.processesList = append(*c.processesList, p)
	return nil
}

func (c *Config) GetProcess(name string) (Process, bool) {
	p, ok := c.processesByName[name]
	return p, ok
}

func (c *Config) DeleteProcess(name string) error {
	if _, ok := c.processesByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "process", Name: name}
	}
	for _, s := range c.ListServices() {
		if s.Process.Name == name {
			return &railgunerrors.ErrDependent{DependentType: "service", DependentName: s.Name, Type: "process", Name: name}
		}
	}
	delete(c.processesByName, name)
	existing := *c.processesList
	keep := make([]Process, len(existing)-1)
	for _, p := range existing {
		if p.Name != name {
			keep = append(keep, p)
		}
	}
	c.processesList = &keep
	return nil
}

func (c *Config) ListProcesses() []Process {
	return *c.processesList
}

func (c *Config) ParseService(obj interface{}) (Service, error) {
	name := gtg.TryGetString(obj, "name", "")
	if len(name) == 0 {
		return Service{}, &railgunerrors.ErrMissingRequiredParameter{Name: "name"}
	}
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	datastore, found := c.GetDataStore(gtg.TryGetString(obj, "datastore", ""))
	if !found {
		return Service{}, errors.Wrap(errors.New("could not find datastore with name "+gtg.TryGetString(obj, "datastore", "")), (&railgunerrors.ErrInvalidConfig{Name: "service", Value: obj}).Error())
	}
	process, found := c.GetProcess(gtg.TryGetString(obj, "process", ""))
	if !found {
		return Service{}, errors.Wrap(errors.New("could not find process with name "+gtg.TryGetString(obj, "process", "")), (&railgunerrors.ErrInvalidConfig{Name: "process", Value: obj}).Error())
	}
	defaults, err := c.ParseMap(obj, "defaults")
	if err != nil {
		return Service{}, errors.Wrap(err, (&railgunerrors.ErrInvalidConfig{Name: "process", Value: obj}).Error())
	}
	s := Service{
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
		DataStore:   datastore,
		Process:     process,
		Defaults:    defaults,
	}
	return s, nil
}

func (c *Config) AddService(s Service) error {
	if _, ok := c.servicesByName[s.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "service", Value: s.Name}
	}
	c.servicesByName[s.Name] = s
	*c.servicesList = append(*c.servicesList, s)
	return nil
}

func (c *Config) GetService(name string) (Service, bool) {
	p, ok := c.servicesByName[name]
	return p, ok
}

func (c *Config) DeleteService(name string) error {
	if _, ok := c.servicesByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "service", Name: name}
	}
	for _, j := range c.ListJobs() {
		if j.Service.Name == name {
			return &railgunerrors.ErrDependent{DependentType: "job", DependentName: j.Name, Type: "service", Name: name}
		}
	}
	delete(c.servicesByName, name)
	existing := *c.servicesList
	keep := make([]Service, len(existing)-1)
	for _, s := range existing {
		if s.Name != name {
			keep = append(keep, s)
		}
	}
	c.servicesList = &keep
	return nil
}

func (c *Config) ListServices() []Service {
	return *c.servicesList
}

func (c *Config) ParseJob(obj interface{}) (Job, error) {
	name := gtg.TryGetString(obj, "name", "")
	title := gtg.TryGetString(obj, "title", "")
	description := gtg.TryGetString(obj, "description", "")
	serviceName := gtg.TryGetString(obj, "service", "")
	if len(serviceName) == 0 {
		return Job{}, &railgunerrors.ErrMissingRequiredParameter{Name: "service"}
	}
	service, found := c.GetService(serviceName)
	if !found {
		return Job{}, &railgunerrors.ErrMissingObject{Type: "service", Name: serviceName}
	}
	variables, err := c.ParseMap(obj, "variables")
	if err != nil {
		return Job{}, errors.Wrap(err, (&railgunerrors.ErrInvalidConfig{Name: "job", Value: obj}).Error())
	}
	j := Job{
		Name:        name,
		Title:       c.Coalesce(title, name),
		Description: c.Coalesce(description, title, name),
		Service:     service,
		Variables:   variables,
	}
	return j, nil
}

func (c *Config) AddJob(j Job) error {
	if _, ok := c.jobsByName[j.Name]; ok {
		return &railgunerrors.ErrAlreadyExists{Name: "job", Value: j.Name}
	}
	c.jobsByName[j.Name] = j
	*c.jobsList = append(*c.jobsList, j)
	return nil
}

func (c *Config) GetJob(name string) (Job, bool) {
	p, ok := c.jobsByName[name]
	return p, ok
}

func (c *Config) DeleteJob(name string) error {
	if _, ok := c.jobsByName[name]; !ok {
		return &railgunerrors.ErrMissingObject{Type: "job", Name: name}
	}
	delete(c.jobsByName, name)
	existing := *c.jobsList
	keep := make([]Job, len(existing)-1)
	for _, j := range existing {
		if j.Name != name {
			keep = append(keep, j)
		}
	}
	c.jobsList = &keep
	return nil
}

func (c *Config) ListJobs() []Job {
	return *c.jobsList
}
