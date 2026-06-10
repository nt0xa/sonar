package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) HTTPRoutesList(w http.ResponseWriter, r *http.Request) {
	routes, err := api.svc.HTTPRoutesList(r.Context(), service.HTTPRoutesListInput{
		PayloadName: r.PathValue("payload"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, routes)
}
