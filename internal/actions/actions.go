package actions

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/utils/errors"
)

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
