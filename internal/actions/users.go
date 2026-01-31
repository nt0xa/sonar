package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/utils/errors"
)

const (
	UsersCreateResultID = "users/create"
	UsersDeleteResultID = "users/delete"
)

type UsersActions interface {
	UsersCreate(context.Context, UsersCreateParams) (*UsersCreateResult, errors.Error)
	UsersDelete(context.Context, UsersDeleteParams) (*UsersDeleteResult, errors.Error)
}

type User struct {
	Name       string    `json:"name"`
	IsAdmin    bool      `json:"isAdmin"`
	APIToken   *string   `json:"apiToken,omitempty"`
	TelegramID *int64    `json:"telegramId,omitempty"`
	LarkID     *string   `json:"larkId,omitempty"`
	SlackID    *string   `json:"slackId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

//
// Create
//

type UsersCreateParams struct {
	Name       string  `json:"name"`
	APIToken   *string `json:"apiToken,omitempty"`
	TelegramID *int64  `json:"telegramId,omitempty"`
	LarkID     *string `json:"larkId,omitempty"`
	SlackID    *string `json:"slackId,omitempty"`
	IsAdmin    bool    `json:"isAdmin"`
}

func (p UsersCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type UsersCreateResult struct {
	User
}

func (r UsersCreateResult) ResultID() string {
	return UsersCreateResultID
}

func UsersCreateCommand(acts *Actions, p *UsersCreateParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new user",
		Long:  "Create new user identified by NAME",
		Args:  oneArg("NAME"),
	}

	var (
		apiToken   string
		telegramID int64
		larkID     string
		slackID    string
	)

	cmd.Flags().BoolVarP(&p.IsAdmin, "admin", "a", false, "Admin user")
	cmd.Flags().StringVar(&apiToken, "token", "", "API token")
	cmd.Flags().Int64Var(&telegramID, "telegram", 0, "Telegram user ID")
	cmd.Flags().StringVar(&larkID, "lark", "", "Lark user ID")
	cmd.Flags().StringVar(&slackID, "slack", "", "Slack user ID")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]

		if cmd.Flags().Changed("token") {
			p.APIToken = &apiToken
		}

		if cmd.Flags().Changed("telegram") {
			p.TelegramID = &telegramID
		}

		if cmd.Flags().Changed("lark") {
			p.LarkID = &larkID
		}

		if cmd.Flags().Changed("slack") {
			p.SlackID = &slackID
		}

		return nil
	}
}

//
// Delete
//

type UsersDeleteParams struct {
	Name string `path:"name"`
}

func (p UsersDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type UsersDeleteResult struct {
	User
}

func (r UsersDeleteResult) ResultID() string {
	return UsersDeleteResultID
}

func UsersDeleteCommand(acts *Actions, p *UsersDeleteParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete user",
		Long:  "Delete user identified by NAME",
		Args:  oneArg("NAME"),
	}

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]

		return nil
	}
}
