package httperr

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}
