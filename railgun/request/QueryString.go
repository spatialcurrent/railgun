// =================================================================
//
// Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
// Released as open source under the MIT License.  See LICENSE file.
//
// =================================================================

package request

import (
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

type QueryString struct {
	Params map[string][]string
}

func (qs QueryString) FirstBool(name string) (bool, error) {
	v, ok := qs.Params[name]
	if !ok {
		return false, &ErrQueryStringParameterMissing{Name: name}
	}
	if len(v) == 0 {
		return false, errors.New("query string parameter " + name + " is empty")
	}
	str := strings.ToLower(v[0])
	return str == "true" || str == "1" || str == "t" || str == "y", nil
}

func (qs QueryString) FirstString(name string) (string, error) {
	v, ok := qs.Params[name]
	if !ok {
		return "", &ErrQueryStringParameterMissing{Name: name}
	}
	if len(v) == 0 {
		return "", errors.New("query string parameter " + name + " is empty")
	}
	return v[0], nil
}

func (qs QueryString) FirstInt(name string) (int, error) {
	s, err := qs.FirstString(name)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.Wrap(err, "query string parameter "+name+" is not an int ("+s+")")
	}
	return i, nil
}

func NewQueryString(r *http.Request) QueryString {
	return QueryString{Params: r.URL.Query()}
}
