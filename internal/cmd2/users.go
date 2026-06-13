package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) usersCreate(cmd *cobra.Command) cmdx.RunFunc {
	var (
		in         service.UsersCreateInput
		apiToken   string
		telegramID int64
		larkID     string
		slackID    string
	)

	cmd.Use = "new NAME"
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().BoolVarP(&in.IsAdmin, "admin", "a", false, "Admin user")
	cmd.Flags().StringVar(&apiToken, "token", "", "API token")
	cmd.Flags().Int64Var(&telegramID, "telegram", 0, "Telegram user ID")
	cmd.Flags().StringVar(&larkID, "lark", "", "Lark user ID")
	cmd.Flags().StringVar(&slackID, "slack", "", "Slack user ID")

	return func(cmd *cobra.Command, args []string) error {
		in.Name = args[0]

		if cmd.Flags().Changed("token") {
			in.APIToken = &apiToken
		}
		if cmd.Flags().Changed("telegram") {
			in.TelegramID = &telegramID
		}
		if cmd.Flags().Changed("lark") {
			in.LarkID = &larkID
		}
		if cmd.Flags().Changed("slack") {
			in.SlackID = &slackID
		}

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.UsersCreate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) usersDelete(cmd *cobra.Command) cmdx.RunFunc {
	var in service.UsersDeleteInput

	cmd.Use = "del NAME"
	cmd.Args = cobra.ExactArgs(1)

	return func(cmd *cobra.Command, args []string) error {
		in.Name = args[0]

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.UsersDelete(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
