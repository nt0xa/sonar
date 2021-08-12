package actions

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
)

type PayloadsActions interface {
	PayloadsCreate(context.Context, PayloadsCreateParams) (PayloadsCreateResult, errors.Error)
	PayloadsUpdate(context.Context, PayloadsUpdateParams) (PayloadsUpdateResult, errors.Error)
	PayloadsDelete(context.Context, PayloadsDeleteParams) (PayloadsDeleteResult, errors.Error)
	PayloadsList(context.Context, PayloadsListParams) (PayloadsListResult, errors.Error)
}

type PayloadsHandler interface {
	PayloadsCreate(context.Context, PayloadsCreateResult)
	PayloadsList(context.Context, PayloadsListResult)
	PayloadsUpdate(context.Context, PayloadsUpdateResult)
	PayloadsDelete(context.Context, PayloadsDeleteResult)
}

type Payload struct {
	Subdomain       string    `json:"subdomain"`
	Name            string    `json:"name"`
	NotifyProtocols []string  `json:"notifyProtocols"`
	StoreEvents     int       `json:"storeEvents"`
	CreatedAt       time.Time `json:"createdAt"`
}

//
// Create
//

type PayloadsCreateParams struct {
	Name            string   `err:"name"            json:"name"`
	NotifyProtocols []string `err:"notifyProtocols" json:"notifyProtocols"`
	StoreEvents     int      `err:"storeEvents"     json:"storeEvents"`
}

func (p PayloadsCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(valid.OneOf(
			models.ProtoCategoriesAll.Strings(),
			true,
		))),
		validation.Field(&p.StoreEvents, validation.Min(0), validation.Max(100)),
	)
}

type PayloadsCreateResult *Payload

func PayloadsCreateCommand(p *PayloadsCreateParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "new NAME",
		Short: "Create new payload",
		Long:  "Create new payload identified by NAME",
		Args:  oneArg("NAME"),
	}

	cmd.Flags().StringSliceVarP(&p.NotifyProtocols, "protocols", "p",
		models.ProtoCategoriesAll.Strings(), "Protocols to notify")
	cmd.Flags().IntVarP(&p.StoreEvents, "events", "e", 0, "Store events in database")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]
		return nil
	}
}

//
// Update
//

type PayloadsUpdateParams struct {
	Name            string   `err:"name"            json:"-"               path:"name"`
	NewName         string   `err:"newName"         json:"name"`
	NotifyProtocols []string `err:"notifyProtocols" json:"notifyProtocols"`
	StoreEvents     int      `err:"storeEvents"     json:"storeEvents"`
}

func (p PayloadsUpdateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.NotifyProtocols, validation.Each(valid.OneOf(
			models.ProtoCategoriesAll.Strings(),
			true,
		))),
		// We need -1 here to find out if the value was changed
		validation.Field(&p.StoreEvents, validation.Min(-1), validation.Max(100)),
	)
}

type PayloadsUpdateResult *Payload

func PayloadsUpdateCommand(p *PayloadsUpdateParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "mod NAME",
		Short: "Modify existing payload",
		Long:  "Modify existing payload identified by NAME",
		Args:  oneArg("NAME"),
	}

	cmd.Flags().StringVarP(&p.NewName, "name", "n", "", "Payload name")
	cmd.Flags().StringSliceVarP(&p.NotifyProtocols, "protocols", "p", nil, "Protocols to notify")
	cmd.Flags().IntVarP(&p.StoreEvents, "events", "e", -1, "Store events in database")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]
		return nil
	}
}

//
// Delete
//

type PayloadsDeleteParams struct {
	Name string `err:"name" path:"name"`
}

func (p PayloadsDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required))
}

type PayloadsDeleteResult *Payload

func PayloadsDeleteCommand(p *PayloadsDeleteParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "del NAME",
		Short: "Delete payload",
		Long:  "Delete payload identified by NAME",
		Args:  oneArg("NAME"),
	}

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Name = args[0]
		return nil
	}
}

//
// List
//

type PayloadsListParams struct {
	Name string `err:"name" query:"name"`
}

func (p PayloadsListParams) Validate() error {
	return nil
}

type PayloadsListResult []*Payload

func PayloadsListCommand(p *PayloadsListParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list SUBSTR",
		Short: "List payloads",
		Long:  "List payloads whose NAME contain SUBSTR",
	}

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		if len(args) > 0 {
			p.Name = args[0]
		}
		return nil
	}
}
