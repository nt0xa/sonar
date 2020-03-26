package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Error struct {
	Code    int    `json:"-"`
	Err     error  `json:"-"`
	Message string `json:"message"`
	Errors  error  `json:"errors,omitempty"`
}

func NewError(code int) *Error {
	return &Error{Code: code, Message: http.StatusText(code)}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s (%v)", e.Code, e.Message, e.Err)
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

func handleError(log logrus.FieldLogger, w http.ResponseWriter, r *http.Request, e *Error) {
	log = log.WithField("uri", r.RequestURI)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Code)
	if err := json.NewEncoder(w).Encode(e); err != nil {
		log.Errorf("Failed to encode JSON: %v", err)
	}
}
