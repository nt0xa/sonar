package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addEvents(g *cmdx.Command) {
	list := &eventsList{c: c}
	g.Add("list", "List payload events", list.run, list.flags)

	get := &eventsGet{c: c}
	g.Add("get", "Get payload event by INDEX", get.run, get.flags)
}

//
// List
//

type eventsList struct {
	c  *Command
	in service.EventsListInput
}

func (x *eventsList) flags(cmd *cobra.Command) {
	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().UintVarP(&x.in.Limit, "limit", "l", 10, "Limit")
	cmd.Flags().UintVarP(&x.in.Offset, "offset", "o", 0, "Offset")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *eventsList) run(cmd *cobra.Command, args []string) error {
	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.EventsList(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Get
//

type eventsGet struct {
	c  *Command
	in service.EventsGetInput
}

func (x *eventsGet) flags(cmd *cobra.Command) {
	cmd.Use = "get INDEX"
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *eventsGet) run(cmd *cobra.Command, args []string) error {
	i, err := parseIndex(args[0])
	if err != nil {
		return err
	}
	x.in.Index = i

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.EventsGet(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
