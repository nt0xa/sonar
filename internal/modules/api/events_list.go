package api

import (
	"cmp"
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

	reverse, err := queryBool(r, "reverse")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	events, err := api.svc.EventsList(r.Context(), service.EventsListInput{
		PayloadName: r.PathValue("payload"),
		// "limit" and "offset" are deprecated, kept for pre-cursor clients.
		Count:   cmp.Or(queryUint(r, "count"), queryUint(r, "limit")),
		After:   after,
		Before:  before,
		Reverse: reverse,
		Offset:  queryUint(r, "offset"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, events)
}
