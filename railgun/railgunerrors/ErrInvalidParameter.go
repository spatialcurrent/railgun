package railgunerrors

import (
	"fmt"
)

type ErrInvalidParameter struct {
	Name  string
	Value interface{}
}

func (e *ErrInvalidParameter) Error() string {
	return "invalid parameter " + e.Name + " with value " + fmt.Sprint(e.Value)
}
