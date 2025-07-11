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
}

func New(db *database.DB, log *slog.Logger, domain string) actions.Actions {
	return &dbactions{db, log, domain}
}
