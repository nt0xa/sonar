package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) PayloadsCreate(w http.ResponseWriter, r *http.Request) {
	var req apimodels.PayloadsCreateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	p, err := api.svc.PayloadsCreate(r.Context(), service.PayloadsCreateInput{
		Name:            req.Name,
		NotifyProtocols: req.NotifyProtocols,
		StoreEvents:     req.StoreEvents,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusCreated, p)
}
