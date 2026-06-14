package api

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nt0xa/sonar/internal/service"
)

type API struct {
	cfg *Config
	log *slog.Logger
	tls *tls.Config

	// ServerService (not plain Service) because the auth middleware needs the
	// AuthContext* identity resolvers.
	svc service.ServerService
}

func New(
	cfg *Config,
	log *slog.Logger,
	tls *tls.Config,
	svc service.ServerService,
) (*API, error) {
	return &API{
		cfg: cfg,
		log: log,
		tls: tls,
		svc: svc,
	}, nil
}

func (api *API) Start() error {
	srv := http.Server{
		Addr:      fmt.Sprintf(":%d", api.cfg.Port),
		Handler:   api.Handler(),
		TLSConfig: api.tls,
	}

	return srv.ListenAndServeTLS("", "")
}
