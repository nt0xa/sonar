package service

import "fmt"

type ErrorKind int

const (
	ErrorKindInternal ErrorKind = iota
	ErrorKindUnauthorized
	ErrorKindConflict
	ErrorKindNotFound
	ErrorKindValidation
)

type Error struct {
	Kind     ErrorKind
	Message  string
	Problems map[string]string // field -> problem; validation only
}

func (e Error) Error() string { return e.Message }

func Unauthorized() Error { return Error{Kind: ErrorKindUnauthorized, Message: "unauthorized"} }

func NotFoundf(format string, a ...any) Error {
	return Error{Kind: ErrorKindNotFound, Message: fmt.Sprintf(format, a...)}
}

func Conflictf(format string, a ...any) Error {
	return Error{Kind: ErrorKindConflict, Message: fmt.Sprintf(format, a...)}
}

func Validation(problems map[string]string) Error {
	return Error{Kind: ErrorKindValidation, Problems: problems}
}
