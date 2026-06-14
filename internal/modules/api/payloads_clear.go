package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) PayloadsClear(w http.ResponseWriter, r *http.Request) {
	payloads, err := api.svc.PayloadsClear(r.Context(), service.PayloadsClearInput{
		Name: r.URL.Query().Get("name"),
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, payloads)
}
