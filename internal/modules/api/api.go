package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/logger"
)

type API struct {
	cfg     *Config
	db      *database.DB
	log     logger.StdLogger
	tls     *tls.Config
	actions actions.Actions
}

func New(cfg *Config, db *database.DB, log logger.StdLogger,
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

	r.Get("/profile", api.ProfileGet)

	r.Route("/payloads", func(r chi.Router) {
		r.Get("/", api.PayloadsList)
		r.Post("/", api.PayloadsCreate)
		r.Delete("/", api.PayloadsClear)
		r.Route("/{name}", func(r chi.Router) {
			r.Delete("/", api.PayloadsDelete)
			r.Put("/", api.PayloadsUpdate)
		})
	})

	r.Route("/dns-records", func(r chi.Router) {
		r.Post("/", api.DNSRecordsCreate)
		r.Route("/{payload}", func(r chi.Router) {
			r.Get("/", api.DNSRecordsList)
			r.Delete("/", api.DNSRecordsClear)
			r.Route("/{index}", func(r chi.Router) {
				r.Delete("/", api.DNSRecordsDelete)
			})
		})
	})

	r.Route("/http-routes", func(r chi.Router) {
		r.Post("/", api.HTTPRoutesCreate)
		r.Route("/{payload}", func(r chi.Router) {
			r.Get("/", api.HTTPRoutesList)
			r.Delete("/", api.HTTPRoutesClear)
			r.Route("/{index}", func(r chi.Router) {
				r.Delete("/", api.HTTPRoutesDelete)
			})
		})
	})

	r.Route("/events", func(r chi.Router) {
		r.Route("/{payload}", func(r chi.Router) {
			r.Get("/", api.EventsList)
			r.Get("/{index}", api.EventsGet)
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
