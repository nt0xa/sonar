package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) DNSRecordsDelete(w http.ResponseWriter, r *http.Request) {
	index, err := pathInt64(r, "index")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	rec, err := api.svc.DNSRecordsDelete(r.Context(), service.DNSRecordsDeleteInput{
		PayloadName: r.PathValue("payload"),
		Index:       index,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, rec)
}
