package errors

import (
	"fmt"
)

type BaseError struct {
	Message string `json:"message"`
}

func (e *BaseError) Error() string {
	return e.Message
}

type InternalError struct {
	BaseError
	Cause error `json:"-"`
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("%s: %+v", e.Message, e.Cause)
}

func Internal(err error) error {
	return &InternalError{
		BaseError: BaseError{
			Message: "internal error",
		},
		Cause: err,
	}
}

func Internalf(format string, args ...interface{}) error {
	return Internal(fmt.Errorf(format, args))
}

type BadFormatError struct {
	BaseError
}

func BadFormatf(format string, args ...interface{}) error {
	return &BadFormatError{
		BaseError: BaseError{
			Message: fmt.Sprintf("bad format: %s", fmt.Sprintf(format, args...)),
		},
	}
}

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
	return fmt.Sprintf("%s: %s", e.Message, e.Errors)
}

func Validation(errs error) error {
	return &ValidationError{
		BaseError: BaseError{
			Message: "validation failed",
		},
		Errors: errs,
	}
}

func Validationf(format string, args ...interface{}) error {
	return &ValidationError{
		BaseError: BaseError{
			Message: fmt.Sprintf("validation failed: %s", fmt.Sprintf(format, args...)),
		},
	}
}

type ConflictError struct {
	BaseError
}

func Conflictf(format string, args ...interface{}) error {
	return &ConflictError{
		BaseError: BaseError{
			Message: fmt.Sprintf("conflict: %s", fmt.Sprintf(format, args...)),
		},
	}
}

type NotFoundError struct {
	BaseError
}

func NotFoundf(format string, args ...interface{}) error {
	return &NotFoundError{
		BaseError: BaseError{
			Message: fmt.Sprintf("not found: %s", fmt.Sprintf(format, args...)),
		},
	}
}

type UnauthorizedError struct {
	BaseError
}

func Unauthorized() error {
	return &UnauthorizedError{
		BaseError: BaseError{
			Message: "unauthorized",
		},
	}
}

func Unauthorizedf(format string, args ...interface{}) error {
	return &UnauthorizedError{
		BaseError: BaseError{
			Message: fmt.Sprintf("unauthorized: %s", fmt.Sprintf(format, args...)),
		},
	}
}

type ForbiddenError struct {
	BaseError
}

func Forbidden() error {
	return &ForbiddenError{
		BaseError: BaseError{
			Message: "access denied",
		},
	}
}

func Forbiddenf(format string, args ...interface{}) error {
	return &ForbiddenError{
		BaseError: BaseError{
			Message: fmt.Sprintf("access denied: %s", fmt.Sprintf(format, args...)),
		},
	}
}
