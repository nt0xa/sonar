package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addUsers(g *cmdx.Command) {
	create := &usersCreate{c: c}
	g.Add("new", "Create new user", create.run, create.flags)

	del := &usersDelete{c: c}
	g.Add("del", "Delete user", del.run, del.flags)
}

//
// Create
//

type usersCreate struct {
	c          *Command
	in         service.UsersCreateInput
	apiToken   string
	telegramID int64
	larkID     string
	slackID    string
}

func (x *usersCreate) flags(cmd *cobra.Command) {
	cmd.Use = "new NAME"
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().BoolVarP(&x.in.IsAdmin, "admin", "a", false, "Admin user")
	cmd.Flags().StringVar(&x.apiToken, "token", "", "API token")
	cmd.Flags().Int64Var(&x.telegramID, "telegram", 0, "Telegram user ID")
	cmd.Flags().StringVar(&x.larkID, "lark", "", "Lark user ID")
	cmd.Flags().StringVar(&x.slackID, "slack", "", "Slack user ID")
}

func (x *usersCreate) run(cmd *cobra.Command, args []string) error {
	x.in.Name = args[0]

	if cmd.Flags().Changed("token") {
		x.in.APIToken = &x.apiToken
	}
	if cmd.Flags().Changed("telegram") {
		x.in.TelegramID = &x.telegramID
	}
	if cmd.Flags().Changed("lark") {
		x.in.LarkID = &x.larkID
	}
	if cmd.Flags().Changed("slack") {
		x.in.SlackID = &x.slackID
	}

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.UsersCreate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Delete
//

type usersDelete struct {
	c  *Command
	in service.UsersDeleteInput
}

func (x *usersDelete) flags(cmd *cobra.Command) {
	cmd.Use = "del NAME"
	cmd.Args = cobra.ExactArgs(1)
}

func (x *usersDelete) run(cmd *cobra.Command, args []string) error {
	x.in.Name = args[0]

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.UsersDelete(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
