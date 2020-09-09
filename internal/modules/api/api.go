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

	r.Get("/user", api.UserCurrent)

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", api.PayloadsList)
		r.Post("/", api.PayloadsCreate)
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.PayloadsDelete)
			r.Put("/", api.PayloadsUpdate)
		})
	})

	r.Route("/dnsrecords", func(r chi.Router) {
		r.Post("/", api.DNSRecordsCreate)
		r.Route("/{payloadName}", func(r chi.Router) {
			r.Get("/", api.DNSRecordsList)
			r.Route("/{name}/{type}", func(r chi.Router) {
				r.Delete("/", api.DNSRecordsDelete)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(api.checkIsAdmin)

		r.Route("/users", func(r chi.Router) {
			r.Post("/", api.UsersCreate)
			r.Route("/{name}", func(r chi.Router) {
				r.Delete("/", api.UsersDelete)
			})
		})
	})

	return r
}
