package railgunerrors

type ErrUnknownImageExtension struct {
	Extension string
}

func (e *ErrUnknownImageExtension) Error() string {
	return "unknown image extension " + e.Extension
}
