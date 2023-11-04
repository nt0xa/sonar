package actions

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/russtone/sonar/internal/utils/errors"
)

func init() {
	validation.ErrorTag = "err"
}

type Actions interface {
	PayloadsActions
	UsersActions
	DNSActions
	ProfileActions
	EventsActions
	HTTPActions
}

type Result interface {
	ResultID() string
}

type PrepareCommandFunc func(*cobra.Command, []string) errors.Error
