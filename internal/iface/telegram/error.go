package telegram

import "fmt"

type Error struct {
	Err      error
	Msg      string
	Internal bool
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func (e *Error) SetError(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) SetMessage(msg string) *Error {
	e.Msg = msg
	return e
}

var (
	ErrInternal           = &Error{Msg: "Internal error", Internal: true}
	ErrUnauthorizedAccess = &Error{Msg: "Unauthorized access"}
	ErrNotFound           = &Error{Msg: "Not found"}
)
