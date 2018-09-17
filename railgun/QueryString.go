package railgun

import (
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type ErrQueryStringParameterNotExist struct {
	Name string
}

func (e ErrQueryStringParameterNotExist) Error() string {
	return "query string parameter " + e.Name + " does not exist"
}

type QueryString struct {
	Params map[string][]string
}

func (qs QueryString) FirstString(name string) (string, error) {
	v, ok := qs.Params[name]
	if !ok {
		return "", &ErrQueryStringParameterNotExist{Name: name}
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
