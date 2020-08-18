package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Command) Users() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
	}

	cmd.AddCommand(c.CreateUser())
	cmd.AddCommand(c.DeleteUser())

	return cmd
}

func (c *Command) CreateUser() *cobra.Command {
	var p actions.CreateUserParams

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new user",
		Long:  "Create new user identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			params, _ := cmd.Flags().GetStringToString("params")
			if err := mapToStruct(params, &p.Params); err != nil {
				return errors.BadFormat(err)
			}

			res, err := c.Actions.CreateUser(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringToStringP("params", "p", map[string]string{}, "User parameters")
	cmd.Flags().BoolVarP(&p.IsAdmin, "admin", "a", false, "Admin user")

	return cmd
}

func (c *Command) DeleteUser() *cobra.Command {
	var p actions.DeleteUserParams

	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete user",
		Long:  "Delete user identified by NAME",
		Args:  OneArg("NAME"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			res, err := c.Actions.DeleteUser(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
