package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/database"
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
	Name      string              `json:"name"`
	Params    database.UserParams `json:"params"`
	IsAdmin   bool                `json:"isAdmin"`
	CreatedAt time.Time           `json:"createdAt"`
}

//
// Create
//

type UsersCreateParams struct {
	Name    string              `err:"name"    json:"name"`
	Params  database.UserParams `err:"params"  json:"params"`
	IsAdmin bool                `err:"isAdmin" json:"isAdmin"`
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

	cmd.Flags().StringToStringP("params", "p", map[string]string{}, "User parameters")
	cmd.Flags().BoolVarP(&p.IsAdmin, "admin", "a", false, "Admin user")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]

		params, _ := cmd.Flags().GetStringToString("params")
		if err := mapToStruct(params, &p.Params); err != nil {
			return errors.BadFormat(err)
		}

		return nil
	}
}

//
// Delete
//

type UsersDeleteParams struct {
	Name string `err:"name" path:"name"`
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
