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
}

type actions struct {
	db  *database.DB
	log logger.StdLogger
}

func New(db *database.DB, log logger.StdLogger) Actions {
	return &actions{db, log}
}
