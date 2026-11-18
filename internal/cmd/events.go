package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func (c *Command) eventsList(cmd *cobra.Command) runFunc {
	var (
		in            service.EventsListInput
		after, before int64
	)

	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().UintVarP(&in.Limit, "limit", "l", 10, "Limit")
	cmd.Flags().UintVarP(&in.Offset, "offset", "o", 0, "Offset")
	cmd.Flags().Int64VarP(&after, "after", "a", 0, "Events with index greater than this, oldest first")
	cmd.Flags().Int64VarP(&before, "before", "b", 0, "Events with index less than this, newest first")

	_ = cmd.MarkFlagRequired("payload")
	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("after") {
			in.After = &after
		}
		if cmd.Flags().Changed("before") {
			in.Before = &before
		}

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

	_ = cmd.MarkFlagRequired("payload")
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
