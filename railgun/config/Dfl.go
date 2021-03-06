package config

import (
	"github.com/pkg/errors"
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-reader-writer/grw"
	"github.com/spatialcurrent/go-simple-serializer/gss"
	"strings"
)

type Dfl struct {
	Expression string `viper:"dfl-expression"`
	Uri        string `viper:"dfl-uri"`
	Vars       string `viper:"dfl-vars"`
}

func (d *Dfl) Variables() (map[string]interface{}, error) {
	dflVars := map[string]interface{}{}
	if len(d.Vars) > 0 {
		_, dflVarsMap, err := dfl.ParseCompileEvaluateMap(
			d.Vars,
			map[string]interface{}{},
			map[string]interface{}{},
			dfl.DefaultFunctionMap,
			dfl.DefaultQuotes)
		if err != nil {
			return dflVars, errors.Wrap(err, "error parsing initial dfl vars as map")
		}
		if m, ok := gss.StringifyMapKeys(dflVarsMap).(map[string]interface{}); ok {
			dflVars = m
		}
	}
	return dflVars, nil
}

func (d Dfl) Node() (dfl.Node, error) {

	expression := d.Expression

	if len(d.Uri) > 0 {
		f, _, err := grw.ReadFromResource(d.Uri, "none", 4096, false, nil)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening dfl file")
		}
		content, err := f.ReadAllAndClose()
		if err != nil {
			return nil, errors.Wrap(err, "Error reading all from dfl file")
		}
		expression = strings.TrimSpace(dfl.RemoveComments(string(content)))
	}

	if len(expression) == 0 {
		return nil, nil
	}

	n, err := dfl.ParseCompile(expression)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing dfl node.")
	}

	return n, nil
}

func (d Dfl) Map() map[string]interface{} {
	return map[string]interface{}{
		"Expression": d.Expression,
		"Uri":        d.Uri,
		"Vars":       d.Vars,
	}
}
