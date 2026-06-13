package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/modules/api2/apimodels"
	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) UsersCreate(w http.ResponseWriter, r *http.Request) {
	var req apimodels.UsersCreateRequest

	if err := api.decodeJSON(r, &req); err != nil {
		api.handleError(w, r, err)
		return
	}

	u, err := api.svc.UsersCreate(r.Context(), service.UsersCreateInput{
		Name:       req.Name,
		APIToken:   req.APIToken,
		TelegramID: req.TelegramID,
		LarkID:     req.LarkID,
		SlackID:    req.SlackID,
		IsAdmin:    req.IsAdmin,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusCreated, u)
}
