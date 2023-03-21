package actions

import (
	"context"

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

type TextResult struct {
	Text string
}

func (s TextResult) ResultID() string {
	return "text"
}

type ErrorResult struct {
	Error errors.Error
}

func (e ErrorResult) ResultID() string {
	return "error"
}

type ResultHandler interface {
	OnResult(context.Context, Result)
}

type PrepareCommandFunc func(*cobra.Command, []string) errors.Error
