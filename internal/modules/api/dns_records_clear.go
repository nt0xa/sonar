package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) DNSRecordsClear(w http.ResponseWriter, r *http.Request) {
	records, err := api.svc.DNSRecordsClear(r.Context(), service.DNSRecordsClearInput{
		PayloadName: r.PathValue("payload"),
		Name:        r.URL.Query().Get("name"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, records)
}
