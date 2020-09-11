package actions

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

func init() {
	validation.ErrorTag = "err"
}

type Actions interface {
	PayloadsActions
	UsersActions
	DNSActions
	UserActions
}

type ResultHandler interface {
	PayloadsHandler
	DNSRecordsHandler
	UsersHandler
	UserHandler
}

type PrepareCommandFunc func(*cobra.Command, []string) errors.Error
