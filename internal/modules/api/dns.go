package api

import (
	"net/http"

	"github.com/bi-zone/sonar/internal/actions"
)

func (api *API) createDNSRecord(w http.ResponseWriter, r *http.Request) {
	var p actions.DNSRecordsCreateParams

	if err := fromJSON(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.DNSRecordsCreate(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusCreated)
}

func (api *API) listDNSRecords(w http.ResponseWriter, r *http.Request) {
	var p actions.DNSRecordsListParams

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.DNSRecordsList(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}

func (api *API) deleteDNSRecord(w http.ResponseWriter, r *http.Request) {
	var p actions.DNSRecordsDeleteParams

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.DNSRecordsDelete(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}
