package actions

import (
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

type MessageResult struct {
	Message string
}

type Actions struct {
	Payloads PayloadsActions
}

type PayloadsActions struct {
	Create CreatePayloadAction
	Delete DeletePayloadAction
	List   ListPayloadsAction
}

type commonDeps struct {
	db  *database.DB
	log logger.StdLogger
}

func New(db *database.DB, log logger.StdLogger) Actions {
	deps := commonDeps{db, log}

	return Actions{
		Payloads: PayloadsActions{
			Create: NewCreatePayloadAction(deps),
			Delete: NewDeletePayloadAction(deps),
			List:   NewListPayloadsAction(deps),
		},
	}
}
