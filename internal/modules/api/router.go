package api

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/bi-zone/sonar/internal/actions"
)

func (api *API) Router() http.Handler {

	r := chi.NewRouter()

	r.Use(api.checkAuth())

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", api.listPayloads)
		r.Post("/", api.createPayload)
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.deletePayload)
		})
	})

	return r
}

func (api *API) createPayload(w http.ResponseWriter, r *http.Request) {
	u, err := getUser(r.Context())
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	var p actions.CreatePayloadParams

	if err := fromJSON(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.CreatePayload(u, p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusCreated)
}

func (api *API) deletePayload(w http.ResponseWriter, r *http.Request) {
	u, err := getUser(r.Context())
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	var p actions.DeletePayloadParams

	if err := fromPath(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	_, err = api.actions.DeletePayload(u, p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) listPayloads(w http.ResponseWriter, r *http.Request) {
	u, err := getUser(r.Context())
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	var p actions.ListPayloadsParams

	if err := fromQuery(r, &p); err != nil {
		handleError(api.log, w, r, err)
		return
	}

	res, err := api.actions.ListPayloads(u, p)
	if err != nil {
		handleError(api.log, w, r, err)
		return
	}

	responseJSON(w, res, http.StatusOK)
}
