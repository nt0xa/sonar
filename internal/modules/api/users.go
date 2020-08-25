package api

import (
	"net/http"

	"github.com/bi-zone/sonar/internal/actions"
)

func (api *API) createUser(w http.ResponseWriter, r *http.Request) {
	var p actions.UsersCreateParams

	if err := fromJSON(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.UsersCreate(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusCreated)
}

func (api *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	var p actions.UsersDeleteParams

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.UsersDelete(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}
