package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/bi-zone/sonar/internal/controller"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/sirupsen/logrus"
)

type API struct {
	cfg       *Config
	db        *database.DB
	log       *logrus.Logger
	tlsConfig *tls.Config
}

var _ controller.Controller = &API{}

func New(cfg *Config, db *database.DB, log *logrus.Logger, tlsConfig *tls.Config) (*API, error) {
	// Set admin API token
	u, err := db.UsersGetByName("admin")

	if err != nil {
		return nil, fmt.Errorf("api: fail to get admin user from db: %w", err)
	}

	u.Params.APIToken = cfg.Admin

	if err := db.UsersUpdate(u); err != nil {
		return nil, fmt.Errorf("api: fail to set admin token in db: %w", err)
	}

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
