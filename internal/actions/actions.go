package actions

import (
	"fmt"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/logger"
)

var (
	ErrParamsCast = errors.Internal(fmt.Errorf("params type cast failed"))
)

type Validatable interface {
	Validate() error
}

type Action struct {
	Params  Validatable
	Execute func(*database.User, interface{}) (interface{}, error)
}

type MessageResult struct {
	Message string
}

type Actions struct {
	Payloads PayloadsActions
}

type PayloadsActions struct {
	Create *Action
	Delete *Action
	List   *Action
}

func New(db *database.DB, log logger.StdLogger) *Actions {
	return &Actions{
		Payloads: PayloadsActions{
			Create: CreatePayloadAction(db, log),
			Delete: DeletePayloadAction(db, log),
			List:   ListPayloadsAction(db, log),
		},
	}
}
