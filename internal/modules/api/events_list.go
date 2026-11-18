package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) EventsList(w http.ResponseWriter, r *http.Request) {
	after, err := queryInt64Ptr(r, "after")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	before, err := queryInt64Ptr(r, "before")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	events, err := api.svc.EventsList(r.Context(), service.EventsListInput{
		PayloadName: r.PathValue("payload"),
		Limit:       queryUint(r, "limit"),
		Offset:      queryUint(r, "offset"),
		After:       after,
		Before:      before,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, events)
}
