package actions

import (
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type Validatable interface {
	Validate() error
}

type Action struct {
	Params  Validatable
	Execute func(*database.User, interface{}) (*ActionResult, error)
}

type ActionResult struct {
	Message string
	Data    interface{}
}

type Actions map[string]*Action

func New(db *database.DB, log logger.StdLogger) Actions {
	return map[string]*Action{
		"new":  NewPayloadAction(db, log),
		"del":  DeletePayloadAction(db, log),
		"list": ListPayloadsAction(db, log),
	}
}
