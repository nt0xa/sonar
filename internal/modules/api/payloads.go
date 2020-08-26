package api

import (
	"net/http"

	"github.com/bi-zone/sonar/internal/actions"
)

func (api *API) createPayload(w http.ResponseWriter, r *http.Request) {
	var p actions.PayloadsCreateParams

	if err := fromJSON(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.PayloadsCreate(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusCreated)
}

func (api *API) listPayloads(w http.ResponseWriter, r *http.Request) {
	var p actions.PayloadsListParams

	if err := fromQuery(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.PayloadsList(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}

func (api *API) updatePayload(w http.ResponseWriter, r *http.Request) {
	var p actions.PayloadsUpdateParams

	if err := fromJSON(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.PayloadsUpdate(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}

func (api *API) deletePayload(w http.ResponseWriter, r *http.Request) {
	var p actions.PayloadsDeleteParams

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.PayloadsDelete(r.Context(), p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}
