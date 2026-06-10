package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) DNSRecordsList(w http.ResponseWriter, r *http.Request) {
	records, err := api.svc.DNSRecordsList(r.Context(), service.DNSRecordsListInput{
		PayloadName: r.PathValue("payload"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, records)
}
