package errors

import (
	"fmt"
)

type Error interface {
	Message() string
	Error() string
}

type BaseError struct {
	Msg string `json:"message"`
	Det string `json:"details,omitempty"`
}

func (e *BaseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.Det)
}

func (e *BaseError) Message() string {
	return e.Msg
}

func (e *BaseError) Details() string {
	return e.Det
}

//
// Internal
//

type InternalError struct {
	BaseError
	Cause error `json:"-"`
}

func Internal(err error) Error {
	return &InternalError{
		BaseError: BaseError{
			Msg: "internal error",
		},
		Cause: err,
	}
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.Cause)
}

func Internalf(format string, args ...interface{}) Error {
	return Internal(fmt.Errorf(format, args...))
}

//
// Bad format
//

type BadFormatError struct {
	BaseError
}

func BadFormatf(format string, args ...interface{}) Error {
	return &BadFormatError{
		BaseError: BaseError{
			Msg: "bad format",
			Det: fmt.Sprintf(format, args...),
		},
	}
}

//
// Validation
//

type Errors map[string]error

func (e Errors) Error() string {
	s := ""
	for field, err := range e {
		s += fmt.Sprintf("%s: %+v;", field, err)
	}
	return s
}

type ValidationError struct {
	BaseError
	Errors error `json:"errors,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Errors != nil {
		return fmt.Sprintf("%s: %s", e.Msg, e.Errors)
	}

	return e.BaseError.Error()
}

func Validation(errs error) Error {
	return &ValidationError{
		BaseError: BaseError{
			Msg: "validation failed",
		},
		Errors: errs,
	}
}

func Validationf(format string, args ...interface{}) Error {
	return &ValidationError{
		BaseError: BaseError{
			Msg: "validation failed",
			Det: fmt.Sprintf(format, args...),
		},
	}
}

//
// Conflict
//

type ConflictError struct {
	BaseError
}

func Conflictf(format string, args ...interface{}) Error {
	return &ConflictError{
		BaseError: BaseError{
			Msg: "conflict",
			Det: fmt.Sprintf(format, args...),
		},
	}
}

//
// NotFound
//

type NotFoundError struct {
	BaseError
}

func NotFoundf(format string, args ...interface{}) Error {
	return &NotFoundError{
		BaseError: BaseError{
			Msg: "not found",
			Det: fmt.Sprintf(format, args...),
		},
	}
}

//
// Unauthorized
//

type UnauthorizedError struct {
	BaseError
}

func Unauthorized() Error {
	return &UnauthorizedError{
		BaseError: BaseError{
			Msg: "unauthorized",
		},
	}
}

func Unauthorizedf(format string, args ...interface{}) Error {
	return &UnauthorizedError{
		BaseError: BaseError{
			Msg: "unauthorized",
			Det: fmt.Sprintf(format, args...),
		},
	}
}

//
// Forbidden
//

type ForbiddenError struct {
	BaseError
}

func Forbidden() Error {
	return &ForbiddenError{
		BaseError: BaseError{
			Msg: "forbidden",
		},
	}
}

func Forbiddenf(format string, args ...interface{}) Error {
	return &ForbiddenError{
		BaseError: BaseError{
			Msg: "forbidden",
			Det: fmt.Sprintf(format, args...),
		},
	}
}
