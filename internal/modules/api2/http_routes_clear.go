package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) HTTPRoutesClear(w http.ResponseWriter, r *http.Request) {
	routes, err := api.svc.HTTPRoutesClear(r.Context(), service.HTTPRoutesClearInput{
		PayloadName: r.PathValue("payload"),
		Path:        r.URL.Query().Get("path"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, routes)
}
