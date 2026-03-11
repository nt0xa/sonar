package actionsdb

import (
	"log/slog"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
)

type dbactions struct {
	db     *database.DB
	log    *slog.Logger
	domain string
	audit  bool
}

func New(db *database.DB, log *slog.Logger, domain string, auditEnabled bool) actions.Actions {
	return &dbactions{db: db, log: log, domain: domain, audit: auditEnabled}
}
