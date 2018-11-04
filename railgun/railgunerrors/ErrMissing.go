package railgunerrors

type ErrMissing struct {
	Type string
	Name string
}

func (e *ErrMissing) Error() string {
	return e.Type + " with name " + e.Name + " is missing"
}
