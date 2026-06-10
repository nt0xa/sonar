package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) EventsList(w http.ResponseWriter, r *http.Request) {
	events, err := api.svc.EventsList(r.Context(), service.EventsListInput{
		PayloadName: r.PathValue("payload"),
		Limit:       queryUint(r, "limit"),
		Offset:      queryUint(r, "offset"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, events)
}
