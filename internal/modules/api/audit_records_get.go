package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) AuditRecordsGet(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	rec, err := api.svc.AuditRecordsGet(r.Context(), service.AuditRecordsGetInput{
		ID: id,
	})
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	api.encodeJSON(w, http.StatusOK, rec)
}
