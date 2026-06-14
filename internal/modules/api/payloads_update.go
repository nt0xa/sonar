package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) PayloadsUpdate(w http.ResponseWriter, r *http.Request) {
	var req apimodels.PayloadsUpdateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	p, err := api.svc.PayloadsUpdate(r.Context(), service.PayloadsUpdateInput{
		Name:            r.PathValue("name"),
		NewName:         req.Name,
		NotifyProtocols: req.NotifyProtocols,
		StoreEvents:     req.StoreEvents,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, p)
}
