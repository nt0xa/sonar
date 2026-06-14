package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) HTTPRoutesDelete(w http.ResponseWriter, r *http.Request) {
	index, err := pathInt64(r, "index")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	route, err := api.svc.HTTPRoutesDelete(r.Context(), service.HTTPRoutesDeleteInput{
		PayloadName: r.PathValue("payload"),
		Index:       index,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, route)
}
