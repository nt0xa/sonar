package cmd2

import (
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addPayloads(g *cmdx.Command) {
	create := &payloadsCreate{c: c}
	g.Add("new", "Create a new payload", create.run, create.flags)

	list := &payloadsList{c: c}
	g.Add("list", "List payloads", list.run, list.flags)

	update := &payloadsUpdate{c: c}
	g.Add("mod", "Modify existing payload", update.run, update.flags)

	del := &payloadsDelete{c: c}
	g.Add("del", "Delete payload", del.run, del.flags)

	clear := &payloadsClear{c: c}
	g.Add("clr", "Delete multiple payloads", clear.run, clear.flags)
}

//
// Create
//

type payloadsCreate struct {
	c  *Command
	in service.PayloadsCreateInput
}

func (x *payloadsCreate) flags(cmd *cobra.Command) {
	cmd.Use = "new NAME"
	cmd.Args = cobra.ExactArgs(1)

	x.in.NotifyProtocols = service.ProtoCategoryValues()

	cmd.Flags().VarP(&protoSlice{p: &x.in.NotifyProtocols}, "protocols", "p", "Protocols to notify")
	cmd.Flags().BoolVarP(&x.in.StoreEvents, "events", "e", false, "Store events in database")

	_ = cmd.RegisterFlagCompletionFunc("protocols", completeMany(service.ProtoCategoryNames()))
}

func (x *payloadsCreate) run(cmd *cobra.Command, args []string) error {
	x.in.Name = args[0]

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.PayloadsCreate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// List
//

type payloadsList struct {
	c  *Command
	in service.PayloadsListInput
}

func (x *payloadsList) flags(cmd *cobra.Command) {
	cmd.Use = "list [SUBSTR]"
	cmd.Args = cobra.MaximumNArgs(1)

	cmd.Flags().UintVarP(&x.in.Page, "page", "p", 1, "Page")
	cmd.Flags().UintVarP(&x.in.PerPage, "per-page", "s", 10, "Per page")
}

func (x *payloadsList) run(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		x.in.Name = args[0]
	}

	out, err := x.c.svc.PayloadsList(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Update
//

type payloadsUpdate struct {
	c           *Command
	in          service.PayloadsUpdateInput
	storeEvents bool
}

func (x *payloadsUpdate) flags(cmd *cobra.Command) {
	cmd.Use = "mod NAME"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = x.c.completePayloadName

	cmd.Flags().StringVarP(&x.in.NewName, "name", "n", "", "Payload name")
	cmd.Flags().VarP(&protoSlice{p: &x.in.NotifyProtocols}, "protocols", "p", "Protocols to notify")
	cmd.Flags().BoolVarP(&x.storeEvents, "events", "e", false, "Store events in database")

	_ = cmd.RegisterFlagCompletionFunc("protocols", completeMany(service.ProtoCategoryNames()))
}

func (x *payloadsUpdate) run(cmd *cobra.Command, args []string) error {
	x.in.Name = args[0]

	if cmd.Flags().Changed("events") {
		x.in.StoreEvents = &x.storeEvents
	}

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.PayloadsUpdate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Delete
//

type payloadsDelete struct {
	c  *Command
	in service.PayloadsDeleteInput
}

func (x *payloadsDelete) flags(cmd *cobra.Command) {
	cmd.Use = "del NAME"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = x.c.completePayloadName
}

func (x *payloadsDelete) run(cmd *cobra.Command, args []string) error {
	x.in.Name = args[0]

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.PayloadsDelete(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Clear
//

type payloadsClear struct {
	c  *Command
	in service.PayloadsClearInput
}

func (x *payloadsClear) flags(cmd *cobra.Command) {
	cmd.Use = "clr [SUBSTR]"
	cmd.Args = cobra.MaximumNArgs(1)
}

func (x *payloadsClear) run(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		x.in.Name = args[0]
	}

	out, err := x.c.svc.PayloadsClear(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
