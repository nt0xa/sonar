package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func (c *Command) payloadsCreate(cmd *cobra.Command) runFunc {
	var in service.PayloadsCreateInput

	cmd.Use = "new NAME"
	cmd.Args = cobra.ExactArgs(1)

	in.NotifyProtocols = service.ProtoCategoryValues()

	cmd.Flags().VarP(&protoSlice{p: &in.NotifyProtocols}, "protocols", "p", "Protocols to notify")
	cmd.Flags().BoolVarP(&in.StoreEvents, "events", "e", false, "Store events in database")

	_ = cmd.RegisterFlagCompletionFunc("protocols", completeMany(service.ProtoCategoryNames()))

	return func(cmd *cobra.Command, args []string) error {
		in.Name = args[0]

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.PayloadsCreate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) payloadsList(cmd *cobra.Command) runFunc {
	var in service.PayloadsListInput

	cmd.Use = "list [SUBSTR]"
	cmd.Args = cobra.MaximumNArgs(1)

	cmd.Flags().UintVarP(&in.Page, "page", "p", 1, "Page")
	cmd.Flags().UintVarP(&in.PerPage, "per-page", "s", 10, "Per page")

	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			in.Name = args[0]
		}

		out, err := c.svc.PayloadsList(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) payloadsUpdate(cmd *cobra.Command) runFunc {
	var (
		in          service.PayloadsUpdateInput
		storeEvents bool
	)

	cmd.Use = "mod NAME"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = c.completePayloadName

	cmd.Flags().StringVarP(&in.NewName, "name", "n", "", "Payload name")
	cmd.Flags().VarP(&protoSlice{p: &in.NotifyProtocols}, "protocols", "p", "Protocols to notify")
	cmd.Flags().BoolVarP(&storeEvents, "events", "e", false, "Store events in database")

	_ = cmd.RegisterFlagCompletionFunc("protocols", completeMany(service.ProtoCategoryNames()))

	return func(cmd *cobra.Command, args []string) error {
		in.Name = args[0]

		if cmd.Flags().Changed("events") {
			in.StoreEvents = &storeEvents
		}

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.PayloadsUpdate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) payloadsDelete(cmd *cobra.Command) runFunc {
	var in service.PayloadsDeleteInput

	cmd.Use = "del NAME"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = c.completePayloadName

	return func(cmd *cobra.Command, args []string) error {
		in.Name = args[0]

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.PayloadsDelete(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) payloadsClear(cmd *cobra.Command) runFunc {
	var in service.PayloadsClearInput

	cmd.Use = "clr [SUBSTR]"
	cmd.Args = cobra.MaximumNArgs(1)

	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			in.Name = args[0]
		}

		out, err := c.svc.PayloadsClear(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
