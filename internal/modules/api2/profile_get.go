package api2

import (
	"net/http"
	"time"
)

type ProfileGetResponse struct {
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`

	APIToken   *string `json:"apiToken,omitempty"`
	TelegramID *int64  `json:"telegramId,omitempty"`
	LarkID     *string `json:"larkId,omitempty"`
	SlackID    *string `json:"slackId,omitempty"`
}

func (api *API) ProfileGet(w http.ResponseWriter, r *http.Request) {
	u, err := api.svc.ProfileGet(r.Context())
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, ProfileGetResponse(*u))
}
