package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) DNSRecordsCreate(w http.ResponseWriter, r *http.Request) {
	var req apimodels.DNSRecordsCreateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	rec, err := api.svc.DNSRecordsCreate(r.Context(), service.DNSRecordsCreateInput{
		PayloadName: req.PayloadName,
		Name:        req.Name,
		TTL:         req.TTL,
		Type:        req.Type,
		Values:      req.Values,
		Strategy:    req.Strategy,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, rec)
}
