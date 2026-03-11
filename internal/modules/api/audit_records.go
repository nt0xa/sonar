package api

import (
	"net/http"

	"github.com/nt0xa/sonar/internal/actions"
)

func (api *API) AuditRecordsList(w http.ResponseWriter, r *http.Request) {
	var params actions.AuditRecordsListParams

	if err := fromQuery(r, &params); err != nil {
		api.handleError(w, r, err)
		return
	}

	res, err := api.actions.AuditRecordsList(r.Context(), params)
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}

func (api *API) AuditRecordsGet(w http.ResponseWriter, r *http.Request) {
	var params actions.AuditRecordsGetParams

	if err := fromPath(r, &params); err != nil {
		api.handleError(w, r, err)
		return
	}

	res, err := api.actions.AuditRecordsGet(r.Context(), params)
	if err != nil {
		api.handleError(w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}
