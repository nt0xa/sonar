package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/controller"
	"github.com/bi-zone/sonar/internal/database"
)

type API struct {
	cfg       *Config
	db        *database.DB
	log       *logrus.Logger
	tlsConfig *tls.Config
}

var _ controller.Controller = &API{}

func New(cfg *Config, db *database.DB, log *logrus.Logger, tlsConfig *tls.Config) (*API, error) {
	return &API{cfg, db, log, tlsConfig}, nil
}

func (a *API) Start() error {
	srv := http.Server{
		Addr:      fmt.Sprintf(":%d", a.cfg.Port),
		Handler:   a.Router(),
		TLSConfig: a.tlsConfig,
	}

	return srv.ListenAndServeTLS("", "")
}
