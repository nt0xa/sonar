package api

import (
	"net/http"

	"github.com/bi-zone/sonar/internal/database/dbactions"
)

func (api *API) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := dbactions.GetUser(r.Context())
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, dbactions.User(u), http.StatusOK)
}
