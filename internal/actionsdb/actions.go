package actionsdb

import (
	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type dbactions struct {
	db     *database.DB
	log    logger.StdLogger
	domain string
}

func New(db *database.DB, log logger.StdLogger, domain string) actions.Actions {
	return &dbactions{db, log, domain}
}
