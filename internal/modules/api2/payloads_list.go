package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) PayloadsList(w http.ResponseWriter, r *http.Request) {
	payloads, err := api.svc.PayloadsList(r.Context(), service.PayloadsListInput{
		Name:    r.URL.Query().Get("name"),
		Page:    queryUint(r, "page"),
		PerPage: queryUint(r, "perPage"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, payloads)
}
