package api

import (
	"net/http"
)

func (api *API) ProfileGet(w http.ResponseWriter, r *http.Request) {
	u, err := api.svc.ProfileGet(r.Context())
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, u)
}
