package railgunerrors

import (
	"fmt"
)

type ErrAlreadyExists struct {
	Name  string
	Value interface{}
}

func (e *ErrAlreadyExists) Error() string {
	return "config " + e.Name + " with name " + fmt.Sprint(e.Value) + " already exists"
}
