package actions

import (
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type MessageResult struct {
	Message string
}

type Actions interface {
	PayloadsActions
	UsersActions
	DNSActions
}

type actions struct {
	db     *database.DB
	log    logger.StdLogger
	domain string
}

func New(db *database.DB, log logger.StdLogger, domain string) Actions {
	return &actions{db, log, domain}
}
