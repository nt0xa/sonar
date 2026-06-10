package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) EventsGet(w http.ResponseWriter, r *http.Request) {
	index, err := pathInt64(r, "index")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	e, err := api.svc.EventsGet(r.Context(), service.EventsGetInput{
		PayloadName: r.PathValue("payload"),
		Index:       index,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, e)
}
