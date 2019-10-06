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

func (c *RailgunCatalog) LoadFromViper(v *viper.Viper) error {

	//
	// Workspaces
	//

	workspacesByName, err := func(workspaces []string) (map[string]*core.Workspace, error) {
		workspacesByName := map[string]*core.Workspace{}
		for _, str := range workspaces {
			if !strings.HasPrefix(str, "{") {
				return workspacesByName, &rerrors.ErrInvalidConfig{Name: "workspace", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(workspace)
		if err != nil {
			return err
		}
	}

	//
	// Datastores
	//

	datastoresByName, err := func(datastores []string, configSkipErrors bool) (map[string]*core.DataStore, error) {
		datastoresByName := map[string]*core.DataStore{}
		for _, str := range datastores {
			if !strings.HasPrefix(str, "{") {
				_, uriPath := splitter.SplitUri(str)
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
				m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(datastore)
		if err != nil {
			return err
		}
	}

	//
	// Layers
	//

	layersByName, err := func(layers []string) (map[string]*core.Layer, error) {
		layersByName := map[string]*core.Layer{}
		for _, str := range layers {
			if !strings.HasPrefix(str, "{") {
				return layersByName, &rerrors.ErrInvalidConfig{Name: "layer", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(layer)
		if err != nil {
			return err
		}
	}

	//
	// Functions
	//

	functionsByName, err := func(functions []string) (map[string]*core.Function, error) {
		functionsByName := map[string]*core.Function{}
		for _, str := range functions {
			if !strings.HasPrefix(str, "{") {
				return functionsByName, &rerrors.ErrInvalidConfig{Name: "function", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
			if err != nil {
				return functionsByName, &rerrors.ErrInvalidConfig{Name: "function", Value: str}
			}
			f, err := c.ParseFunction(m)
			if err != nil {
				return functionsByName, errors.Wrap(err, "error parsing function")
			}
			functionsByName[f.Name] = f
		}
		return functionsByName, nil
	}(v.GetStringArray("function"))

	for _, function := range functionsByName {
		err = c.Add(function)
		if err != nil {
			return err
		}
	}

	//
	// Processes
	//

	processesByName, err := func(processes []string) (map[string]*core.Process, error) {
		processesByName := map[string]*core.Process{}
		for _, str := range processes {
			if !strings.HasPrefix(str, "{") {
				return processesByName, &rerrors.ErrInvalidConfig{Name: "process", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(process)
		if err != nil {
			return err
		}
	}

	//
	// Services
	//

	servicesByName, err := func(services []string, configSkipErrors bool) (map[string]*core.Service, error) {
		servicesByName := map[string]*core.Service{}
		for _, str := range services {
			if !strings.HasPrefix(str, "{") {
				return servicesByName, &rerrors.ErrInvalidConfig{Name: "service", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(service)
		if err != nil {
			return err
		}
	}

	//
	// Jobs
	//

	jobsByName, err := func(jobs []string, configSkipErrors bool) (map[string]*core.Job, error) {
		jobsByName := map[string]*core.Job{}
		for _, str := range jobs {
			if !strings.HasPrefix(str, "{") {
				return jobsByName, &rerrors.ErrInvalidConfig{Name: "job", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(job)
		if err != nil {
			return err
		}
	}

	//
	// Workflows
	//

	workflowsByName, err := func(workflows []string, configSkipErrors bool) (map[string]*core.Workflow, error) {
		workflowsByName := map[string]*core.Workflow{}
		for _, str := range workflows {
			if !strings.HasPrefix(str, "{") {
				return workflowsByName, &rerrors.ErrInvalidConfig{Name: "workflow", Value: str}
			}
			m, err := parseCompileEvaluateMap(str)
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
		err = c.Add(workflow)
		if err != nil {
			return err
		}
	}

	return nil

}
