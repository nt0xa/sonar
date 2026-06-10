package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) PayloadsDelete(w http.ResponseWriter, r *http.Request) {
	p, err := api.svc.PayloadsDelete(r.Context(), service.PayloadsDeleteInput{
		Name: r.PathValue("name"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, p)
}
