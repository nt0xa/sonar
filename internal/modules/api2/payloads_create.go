package api2

import (
	"net/http"
	"time"

	"github.com/nt0xa/sonar/internal/service"
)

type PayloadsCreateRequest struct {
	Name            string                  `json:"name"`
	NotifyProtocols []service.ProtoCategory `json:"notifyProtocols"`
	StoreEvents     bool                    `json:"storeEvents"`
}

type PayloadsCreateResponse struct {
	Name            string                  `json:"name"`
	Subdomain       string                  `json:"subdomain"`
	NotifyProtocols []service.ProtoCategory `json:"notifyProtocols"`
	StoreEvents     bool                    `json:"storeEvents"`
	CreatedAt       time.Time               `json:"createdAt"`
}

func (api *API) PayloadsCreate(w http.ResponseWriter, r *http.Request) {
	var req PayloadsCreateRequest

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

	api.encodeJSON(w, http.StatusOK, PayloadsCreateResponse{
		Name:            p.Name,
		Subdomain:       p.Subdomain,
		NotifyProtocols: p.NotifyProtocols,
		StoreEvents:     p.StoreEvents,
		CreatedAt:       p.CreatedAt,
	})
}
