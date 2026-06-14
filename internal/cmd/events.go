package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func (c *Command) eventsList(cmd *cobra.Command) runFunc {
	var in service.EventsListInput

	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().UintVarP(&in.Limit, "limit", "l", 10, "Limit")
	cmd.Flags().UintVarP(&in.Offset, "offset", "o", 0, "Offset")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.EventsList(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) eventsGet(cmd *cobra.Command) runFunc {
	var in service.EventsGetInput

	cmd.Use = "get INDEX"
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		i, err := parseIndex(args[0])
		if err != nil {
			return err
		}
		in.Index = i

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.EventsGet(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
