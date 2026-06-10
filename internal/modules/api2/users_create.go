package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

type UsersCreateRequest struct {
	Name       string  `json:"name"`
	APIToken   *string `json:"apiToken"`
	TelegramID *int64  `json:"telegramId"`
	LarkID     *string `json:"larkId"`
	SlackID    *string `json:"slackId"`
	IsAdmin    bool    `json:"isAdmin"`
}

func (api *API) UsersCreate(w http.ResponseWriter, r *http.Request) {
	var req UsersCreateRequest

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

	api.encodeJSON(w, http.StatusOK, u)
}
