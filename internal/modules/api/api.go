package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
)

type API struct {
	cfg     *Config
	db      *database.DB
	log     *logrus.Logger
	tls     *tls.Config
	actions actions.Actions
}

func New(cfg *Config, db *database.DB, log *logrus.Logger,
	tls *tls.Config, actions actions.Actions) (*API, error) {

	return &API{
		cfg:     cfg,
		db:      db,
		log:     log,
		tls:     tls,
		actions: actions,
	}, nil
}

func (api *API) Start() error {
	srv := http.Server{
		Addr:      fmt.Sprintf(":%d", api.cfg.Port),
		Handler:   api.Router(),
		TLSConfig: api.tls,
	}

	return srv.ListenAndServeTLS("", "")
}

func (api *API) Router() http.Handler {

	r := chi.NewRouter()

	r.Use(api.checkAuth())

	r.Get("/user", api.getUser)

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", api.listPayloads)
		r.Post("/", api.createPayload)
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.deletePayload)
			r.Put("/", api.updatePayload)
		})
	})

	r.Route("/dns", func(r chi.Router) {
		r.Post("/", api.createDNSRecord)
		r.Route("/{payloadName}", func(r chi.Router) {
			r.Get("/", api.listDNSRecords)
			r.Route("/{name}/{type}", func(r chi.Router) {
				r.Delete("/", api.deleteDNSRecord)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(api.checkIsAdmin)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", api.createUser)
			r.Route("/{name}", func(r chi.Router) {
				r.Delete("/", api.deleteUser)
			})
		})
	})

	return r
}
