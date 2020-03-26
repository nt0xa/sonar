package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils"
)

func (a *API) Router() http.Handler {

	r := chi.NewRouter()

	r.Use(checkAuth(a.db, a.log))

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", listPayloads(a.db, a.log))
		r.Post("/", createPayload(a.db, a.log))
		r.Route("/{payloadName}", func(r chi.Router) {
			r.Use(setPayload(a.db, a.log))
			r.Delete("/", deletePayload(a.db, a.log))
		})
	})

	return r
}

func listPayloads(db *database.DB, log *logrus.Logger) http.HandlerFunc {

	type Response = []*database.Payload

	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value(userKey).(*database.User)
		if !ok {
			handleError(log, w, r, NewError(500).SetError(errGetUser))
			return
		}

		pp, err := db.PayloadsFindByUserID(u.ID)
		if err != nil {
			handleError(log, w, r, NewError(500).SetError(errGetUser))
			return
		}

		var res Response = pp

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func createPayload(db *database.DB, log *logrus.Logger) http.HandlerFunc {

	type Request struct {
		Name string `json:"name"`
	}

	type Response = *database.Payload

	validate := func(req Request) error {
		return validation.ValidateStruct(&req,
			validation.Field(&req.Name, validation.Required))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		if err, msg := parseJSON(w, r, &req); err != nil {
			handleError(log, w, r, NewError(400).SetMessage(msg))
			return
		}

		if err := validate(req); err != nil {
			handleError(log, w, r, NewError(400).SetErrors(err))
			return
		}

		u, ok := r.Context().Value(userKey).(*database.User)
		if !ok {
			handleError(log, w, r, NewError(500).SetError(errGetUser))
			return
		}

		if _, err := db.PayloadsGetByUserAndName(u.ID, req.Name); err != sql.ErrNoRows {
			handleError(log, w, r, NewError(409).
				SetMessage(fmt.Sprintf("You already have payload with name %q", req.Name)))
			return
		}

		subdomain, err := utils.GenerateRandomString(4)
		if err != nil {
			handleError(log, w, r, NewError(500).SetError(err))
			return
		}

		p := &database.Payload{
			UserID:    u.ID,
			Subdomain: subdomain,
			Name:      req.Name,
		}

		err = db.PayloadsCreate(p)
		if err != nil {
			handleError(log, w, r, NewError(500).SetError(err))
			return
		}

		var res Response = p

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(res)
	}
}

func deletePayload(db *database.DB, log *logrus.Logger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(payloadKey).(*database.Payload)
		if !ok {
			handleError(log, w, r, NewError(500).SetError(errGetPayload))
			return
		}

		err := db.PayloadsDelete(p.ID)
		if err != nil {
			handleError(log, w, r, NewError(500).SetError(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
