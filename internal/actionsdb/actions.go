package actionsdb

import (
	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/logger"
)

type dbactions struct {
	db     *database.DB
	log    logger.StdLogger
	domain string
}

func New(db *database.DB, log logger.StdLogger, domain string) actions.Actions {
	return &dbactions{db, log, domain}
}
