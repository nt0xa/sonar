package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Error struct {
	Code    int    `json:"-"`
	Err     error  `json:"-"`
	Message string `json:"message"`
	Errors  error  `json:"errors"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s (%s)", e.Code, e.Message, e.Err.Error())
}

func (e *Error) SetErrors(errors error) *Error {
	e.Errors = errors
	return e
}

func (e *Error) SetError(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) SetMessage(msg string) *Error {
	e.Message = msg
	return e
}

var (
	ErrForbidden = &Error{Code: http.StatusForbidden, Message: "Access denied"}
)

func HandleError(log logrus.FieldLogger, w http.ResponseWriter, r *http.Request, err error) {

	log = log.WithField("uri", r.RequestURI)

	var (
		e   *Error
		res interface{}
	)

	switch {
	case errors.As(err, &e):
		w.WriteHeader(e.Code)
		res = e

	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Internal server error: %v", err)
		res = &Error{Message: http.StatusText(http.StatusInternalServerError)}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Errorf("Failed to encode JSON: %v", err)
	}
}
