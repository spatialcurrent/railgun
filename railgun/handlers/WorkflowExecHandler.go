// =================================================================
//
// Copyright (C) 2018 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package handlers

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	rerrors "github.com/spatialcurrent/railgun/railgun/errors"
	"github.com/spatialcurrent/railgun/railgun/util"
	"net/http"
	//"reflect"
)

type WorkflowExecHandler struct {
	*BaseHandler
}

func (h *WorkflowExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := util.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "POST":
		obj, err := h.Post(w, r, format, mux.Vars(r))
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		} else {
			err = h.RespondWithObject(w, obj, format)
			if err != nil {
				h.Messages <- err
				err = h.RespondWithError(w, err, format)
				if err != nil {
					panic(err)
				}
			}
		}
	case "OPTIONS":
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *WorkflowExecHandler) Post(w http.ResponseWriter, r *http.Request, format string, vars map[string]string) (interface{}, error) {

	workflowName, ok := vars["name"]
	if !ok {
		return nil, &rerrors.ErrMissingRequiredParameter{Name: "name"}
	}

	fmt.Println("Workflow Name:", workflowName)

	workflow, ok := h.Catalog.GetWorkflow(workflowName)
	if !ok {
		return nil, &rerrors.ErrMissingObject{Type: "workflow", Name: workflowName}
	}

	results := map[string]interface{}{}
	exitCodes := map[string]int{}
	errorBuffers := map[string]*bytes.Buffer{}

	for _, job := range workflow.Jobs {
		errorWriter, errorBuffer := grw.WriteMemoryBytes()
		exitCodes[job.Name] = 0
		errorBuffers[job.Name] = errorBuffer

		if job.Output != nil {
			errorWriter.WriteError(&rerrors.ErrMissingRequiredParameter{Name: "output"})
			exitCodes[job.Name] = 1
			continue
		}

		variables := map[string]interface{}{}
		for k, v := range job.Service.Defaults {
			variables[k] = v
		}
		for k, v := range job.Variables {
			variables[k] = v
		}
		for k, v := range workflow.Variables {
			variables[k] = v
		}

		_, inputUri, err := dfl.EvaluateString(job.Service.DataStore.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "invalid data store uri"))
			exitCodes[job.Name] = 1
			continue
		}

		inputReader, _, err := grw.ReadFromResource(inputUri, job.Service.DataStore.Compression, 4096, false, nil)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error opening resource at uri "+inputUri))
			exitCodes[job.Name] = 1
			continue
		}

		inputBytes, err := inputReader.ReadAllAndClose()
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error reading from resource at uri "+inputUri))
			exitCodes[job.Name] = 1
			continue
		}

		inputFormat := job.Service.DataStore.Format

		inputType, err := gss.GetType(inputBytes, inputFormat)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error getting type for input"))
			exitCodes[job.Name] = 1
			continue
		}

		inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, []string{}, "", false, gss.NoLimit, inputType, false)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error deserializing input using format "+inputFormat))
			exitCodes[job.Name] = 1
			continue
		}

		_, outputObject, err := job.Service.Process.Node.Evaluate(variables, inputObject, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
		if err != nil {
			errorWriter.WriteError(errors.Wrap(err, "error evaluating process with name "+job.Service.Process.Name))
			exitCodes[job.Name] = 1
			continue
		}

		if job.Output != nil {

			outputBytes, err := gss.SerializeBytes(outputObject, job.Output.Format, []string{}, gss.NoLimit)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error serializing output using format "+job.Output.Format))
				exitCodes[job.Name] = 1
				continue
			}

			_, outputUri, err := dfl.EvaluateString(job.Output.Uri, variables, map[string]interface{}{}, dfl.DefaultFunctionMap, dfl.DefaultQuotes)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error evaluating output uri"))
				exitCodes[job.Name] = 1
				continue
			}

			outputWriter, err := grw.WriteToResource(outputUri, job.Output.Compression, false, nil)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error opening output for job "+job.Name))
				exitCodes[job.Name] = 1
				continue
			}

			_, err = outputWriter.Write(outputBytes)
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error writing output for job "+job.Name))
				exitCodes[job.Name] = 1
				continue
			}

			err = outputWriter.Close()
			if err != nil {
				errorWriter.WriteError(errors.Wrap(err, "error closing output for job "+job.Name))
				exitCodes[job.Name] = 1
				continue
			}

		} else {
			results[job.Name] = outputObject
		}

	}

	success := true
	for _, exitCode := range exitCodes {
		if exitCode > 0 {
			success = false
			break
		}
	}

	stderr := map[string]string{}
	for job, buffer := range errorBuffers {
		stderr[job] = buffer.String()
	}

	data := map[string]interface{}{
		"success":   success,
		"message":   "workflow with name " + workflowName + " completed.",
		"exitCodes": exitCodes,
		"stderr":    stderr,
		"results":   results,
	}

	fmt.Println(data)

	return data, nil

}
