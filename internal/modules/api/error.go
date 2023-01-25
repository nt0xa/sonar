package api

import (
	"encoding/json"
	"net/http"

	"github.com/russtone/sonar/internal/utils/errors"
)

func (api *API) handleError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch err.(type) {

	case *errors.BadFormatError, *errors.ValidationError:
		w.WriteHeader(http.StatusBadRequest)

	case *errors.NotFoundError:
		w.WriteHeader(http.StatusNotFound)

	case *errors.ConflictError:
		w.WriteHeader(http.StatusConflict)

	case *errors.UnauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)

	case *errors.ForbiddenError:
		w.WriteHeader(http.StatusForbidden)

	case *errors.InternalError:
		w.WriteHeader(http.StatusInternalServerError)

	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(err); err != nil {
		api.log.Printf("Failed to encode JSON: %v", err)
	}
}
