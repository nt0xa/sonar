package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) UsersDelete(w http.ResponseWriter, r *http.Request) {
	u, err := api.svc.UsersDelete(r.Context(), service.UsersDeleteInput{
		Name: r.PathValue("name"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, u)
}
