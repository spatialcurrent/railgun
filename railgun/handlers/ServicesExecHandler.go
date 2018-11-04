package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"github.com/spatialcurrent/go-try-get/gtg"
	"github.com/spatialcurrent/railgun/railgun"
	"net/http"
	"reflect"
)

type ServicesExecHandler struct {
	*BaseHandler
}

func (h *ServicesExecHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	_, format, _ := railgun.SplitNameFormatCompression(r.URL.Path)

	switch r.Method {
	case "POST":
		b, err := h.Post(w, r, format)
		if err != nil {
			h.Messages <- err
			err = h.RespondWithError(w, err, format)
			if err != nil {
				panic(err)
			}
		}
		w.Write(b)
	default:
		err := h.RespondWithNotImplemented(w, format)
		if err != nil {
			panic(err)
		}
	}

}

func (h *ServicesExecHandler) Post(w http.ResponseWriter, r *http.Request, format string) ([]byte, error) {

	obj, err := h.ParseBody(r, format)
	if err != nil {
		return make([]byte, 0), err
	}

	serviceName := gtg.TryGetString(obj, "service", "")
	if len(serviceName) == 0 {
		return make([]byte, 0), errors.New("missing service name")
	}

	service, ok := h.Config.GetService(serviceName)
	if !ok {
		return make([]byte, 0), errors.New("no service with name " + serviceName)
	}

	vars := map[string]interface{}{}
	for k, v := range service.Defaults {
		vars[k] = v
	}
	input := gtg.TryGetString(obj, "input", "")
	if len(input) > 0 {
		_, m, err := dfl.ParseCompileEvaluateMap(input, map[string]interface{}{}, map[string]interface{}{}, h.DflFuncs, dfl.DefaultQuotes)
		if err != nil {
			fmt.Println(errors.Wrap(err, "invalid input variables"))
			return make([]byte, 0), errors.Wrap(err, "invalid input variables")
		}
		m2 := gss.StringifyMapKeys(m)
		if m3, ok := m2.(map[string]interface{}); ok {
			vars = m3
		}
	}
	
	fmt.Println("Vars:", vars)

	_, inputUri, err := dfl.EvaluateString(service.DataStore.Uri, vars, map[string]interface{}{}, h.DflFuncs, dfl.DefaultQuotes)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "invalid data store uri")
	}

	inputReader, _, err := grw.ReadFromResource(inputUri, service.DataStore.Compression, 4096, false, nil)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error opening resource at uri "+inputUri)
	}

	inputBytes, err := inputReader.ReadAllAndClose()
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error reading from resource at uri "+inputUri)
	}

	inputFormat := service.DataStore.Format

	inputType, err := gss.GetType(inputBytes, inputFormat)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error getting type for input")
	}

	inputObject, err := gss.DeserializeBytes(inputBytes, inputFormat, []string{}, "", false, gss.NoLimit, inputType, false)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error deserializing input using format "+inputFormat)
	}
	
  fmt.Println("Vars:", vars)
  
  fmt.Println("Vars:", vars)

	_, outputObject, err := service.Process.Node.Evaluate(vars, inputObject, h.DflFuncs, dfl.DefaultQuotes)
	if err != nil {
		return make([]byte, 0), errors.Wrap(err, "error evaluatig process with name "+service.Process.Name)
	}
	
	fmt.Println("outputObject:", reflect.TypeOf(outputObject), outputObject)

	b, err := gss.SerializeBytes(outputObject, format, []string{}, gss.NoLimit)
	if err != nil {
		return b, errors.Wrap(err, "error serializing response body")
	}
	return b, nil

}
