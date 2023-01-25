package actionsdb

import (
	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/utils/logger"
)

type dbactions struct {
	db     *database.DB
	log    logger.StdLogger
	domain string
}

func New(db *database.DB, log logger.StdLogger, domain string) actions.Actions {
	return &dbactions{db, log, domain}
}
