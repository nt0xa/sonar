package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
)

type API struct {
	cfg     *Config
	db      *database.DB
	log     *logrus.Logger
	tls     *tls.Config
	actions *actions.Actions

	decoder *schema.Decoder
}

func New(cfg *Config, db *database.DB, log *logrus.Logger,
	tls *tls.Config, actions *actions.Actions) (*API, error) {

	return &API{
		cfg:     cfg,
		db:      db,
		log:     log,
		tls:     tls,
		actions: actions,
		decoder: schema.NewDecoder(),
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
