package api2

import (
	"errors"
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

type httpError struct {
	Status   int               `json:"-"`
	Message  string            `json:"message"`
	Problems map[string]string `json:"problems,omitempty"`
}

func (e httpError) Error() string {
	return e.Message
}

func (api *API) handleError(w http.ResponseWriter, r *http.Request, err error) {
	e := api.toHTTPError(r, err)
	api.encodeJSON(w, e.Status, e)
}

func (api *API) toHTTPError(r *http.Request, err error) httpError {
	if e, ok := errors.AsType[httpError](err); ok { // already shaped (e.g. decodeJSON)
		return e
	}

	se, ok := errors.AsType[service.Error](err)
	if !ok {
		api.logInternal(r, err)
		return httpError{Status: http.StatusInternalServerError, Message: "internal error"}
	}

	switch se.Kind {
	case service.ErrorKindNotFound:
		return httpError{Status: http.StatusNotFound, Message: se.Message}
	case service.ErrorKindConflict:
		return httpError{Status: http.StatusConflict, Message: se.Message}
	case service.ErrorKindUnauthorized:
		return httpError{Status: http.StatusUnauthorized, Message: se.Message}
	case service.ErrorKindValidation:
		return httpError{Status: http.StatusUnprocessableEntity, Message: "validation failed", Problems: se.Problems}
	default:
		api.logInternal(r, err)
		return httpError{Status: http.StatusInternalServerError, Message: "internal error"}
	}
}

func (api *API) logInternal(r *http.Request, err error) {
	api.log.Error("internal error",
		"method", r.Method, "path", r.URL.Path, "error", err)
}
