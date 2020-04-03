package api

import (
	"crypto/tls"

	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/database"
)

type API struct {
	cfg *Config
	db  *database.DB
	log *logrus.Logger
	tls *tls.Config
}

func New(cfg *Config, db *database.DB, log *logrus.Logger, tls *tls.Config) (*API, error) {
	return &API{cfg, db, log, tls}, nil
}
