package railgunerrors

type ErrMissingCollection struct {
	Name string
}

func (e *ErrMissingCollection) Error() string {
	return "missing collection " + e.Name
}
