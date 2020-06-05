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
		r.Get("/", api.listPayloads(api.actions.Payloads.List))
		r.Post("/", api.createPayload(api.actions.Payloads.Create))
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.deletePayload(api.actions.Payloads.Delete))
		})
	})

	return r
}

func (api *API) createPayload(action actions.CreatePayloadAction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		res, err := action.Execute(u, p)
		if err != nil {
			handleError(api.log, w, r, err)
			return
		}

		responseJSON(w, res, http.StatusCreated)
	}
}

func (api *API) deletePayload(action actions.DeletePayloadAction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		_, err = action.Execute(u, p)
		if err != nil {
			handleError(api.log, w, r, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (api *API) listPayloads(action actions.ListPayloadsAction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		res, err := action.Execute(u, p)
		if err != nil {
			handleError(api.log, w, r, err)
			return
		}

		responseJSON(w, res, http.StatusOK)
	}
}
