package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func UsersCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage users",
		PersistentPreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return errors.Internal(err)
			}

			if !u.Params.Admin {
				return errors.Forbidden()
			}

			return nil
		}),
	}

	cmd.AddCommand(CreateUserCmd(acts, handler))
	cmd.AddCommand(DeleteUserCmd(acts, handler))

	return cmd
}

func CreateUserCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var p actions.CreateUserParams

	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new user",
		Long:  "Create new user identified by NAME",
		Args:  OneArg("NAME"),
		PreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			params, _ := cmd.Flags().GetStringToString("params")

			if err := mapToStruct(params, &p.Params); err != nil {
				return errors.Validation(err)
			}

			if err := p.Validate(); err != nil {
				return errors.Validation(err)
			}

			return nil
		}),
		RunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			res, err := acts.CreateUser(u, p)
			if err != nil {
				return err
			}

			handler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringToStringP("params", "p", map[string]string{}, "User parameters")

	return cmd
}

func DeleteUserCmd(acts actions.Actions, handler ResultHandler) *cobra.Command {
	var p actions.DeleteUserParams

	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete user",
		Long:  "Delete user identified by NAME",
		Args:  OneArg("NAME"),
		PreRunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Name = args[0]

			if err := p.Validate(); err != nil {
				return errors.Validation(err)
			}

			return nil
		}),
		RunE: runE(func(cmd *cobra.Command, args []string) errors.Error {
			u, err := GetUser(cmd.Context())
			if err != nil {
				return err
			}

			res, err := acts.DeleteUser(u, p)
			if err != nil {
				return err
			}

			handler(cmd.Context(), res)

			return nil
		}),
	}

	return cmd
}
