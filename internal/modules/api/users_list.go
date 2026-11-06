package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) UsersList(w http.ResponseWriter, r *http.Request) {
	users, err := api.svc.UsersList(r.Context(), service.UsersListInput{
		Name:    r.URL.Query().Get("name"),
		Page:    queryUint(r, "page"),
		PerPage: queryUint(r, "perPage"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, users)
}
