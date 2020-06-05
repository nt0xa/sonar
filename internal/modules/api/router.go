package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/parse"
)

type paramsPlacement int

const (
	inQuery paramsPlacement = iota
	inBody
	inPath
)

func (api *API) Router() http.Handler {

	r := chi.NewRouter()

	r.Use(checkAuth(api.db, api.log))

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", api.handler(api.actions.Payloads.List, inQuery, http.StatusOK))
		r.Post("/", api.handler(api.actions.Payloads.Create, inBody, http.StatusCreated))
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.handler(api.actions.Payloads.Delete, inPath, http.StatusNoContent))
		})
	})

	return r
}

func (api *API) handler(action *actions.Action, params paramsPlacement, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value(userKey).(*database.User)
		if !ok {
			handleError(api.log, w, r, errors.Internalf("no %q key in context", userKey))
			return
		}

		p := action.Params

		if p == nil {
			handleError(api.log, w, r, errors.Internalf("nil params"))
			return
		}

		switch params {

		case inBody:
			rdr := http.MaxBytesReader(w, r.Body, 1024*1024)

			if err := parse.JSON(rdr, p); err != nil {
				handleError(api.log, w, r, errors.BadFormatf("json: %s", err))
				return
			}

		case inQuery:
			if err := api.decoder.Decode(p, r.URL.Query()); err != nil {
				handleError(api.log, w, r, errors.BadFormatf("query: %s", err))
				return
			}

		case inPath:
			if err := pathDecode(r, p); err != nil {
				handleError(api.log, w, r, errors.BadFormatf("path: %s", err))
				return
			}
		}

		if err := p.Validate(); err != nil {
			handleError(api.log, w, r, errors.Validation(err))
			return
		}

		res, err := action.Execute(u, p)
		if err != nil {
			handleError(api.log, w, r, err)
			return
		}

		if res == nil {
			w.WriteHeader(status)
			return
		}

		if _, ok := res.(*actions.MessageResult); ok {
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(res)
	}
}
