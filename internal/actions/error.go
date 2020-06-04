package actions

import (
	"errors"
	"fmt"
)

var (
	castError = ErrInternal(errors.New("type cast failed"))
)

type BaseError struct {
	Message string
}

func (e *BaseError) Error() string {
	return e.Message
}

type InternalError struct {
	BaseError
	Cause error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("%s: %+v", e.Message, e.Cause)
}

func ErrInternal(err error) error {
	return &InternalError{
		BaseError: BaseError{
			Message: "internal error",
		},
		Cause: err,
	}
}

type ValidationError struct {
	BaseError
	Errors map[string]error
}

func (e *ValidationError) Error() string {
	errs := ""
	for field, err := range e.Errors {
		errs += fmt.Sprintf("%s: %+v;", field, err)
	}
	return fmt.Sprintf("%s: %s", e.Message, errs)
}

func ErrValidation(errs map[string]error) error {
	return &ValidationError{
		BaseError: BaseError{
			Message: "validation error",
		},
		Errors: errs,
	}
}

type ConflictError struct {
	BaseError
}

func ErrConflict(format string, args ...interface{}) error {
	return &ValidationError{
		BaseError: BaseError{
			Message: fmt.Sprintf("conflict: %s", fmt.Sprintf(format, args...)),
		},
	}
}

type NotFoundError struct {
	BaseError
}

func ErrNotFound(format string, args ...interface{}) error {
	return &ValidationError{
		BaseError: BaseError{
			Message: fmt.Sprintf("not found: %s", fmt.Sprintf(format, args...)),
		},
	}
}
