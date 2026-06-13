package cmd2

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addAudit(g *cmdx.Command) {
	list := &auditRecordsList{c: c}
	g.Add("list", "List audit records", list.run, list.flags)

	get := &auditRecordsGet{c: c}
	g.Add("get", "Get audit record by ID", get.run, get.flags)
}

//
// List
//

type auditRecordsList struct {
	c       *Command
	in      service.AuditRecordsListInput
	actorID int64
	from    string
	to      string
}

func (x *auditRecordsList) flags(cmd *cobra.Command) {
	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().Int64Var(&x.actorID, "actor-id", 0, "Filter by actor ID")
	cmd.Flags().StringVar(&x.in.ActorName, "actor-name", "", "Filter by actor name")
	cmd.Flags().Var(auditResourceTypeValue{&x.in.ResourceType}, "resource-type", "Filter by resource type")
	cmd.Flags().Var(auditActionValue{&x.in.Action}, "action", "Filter by action")
	cmd.Flags().UintVarP(&x.in.Page, "page", "p", 1, "Page")
	cmd.Flags().UintVarP(&x.in.PerPage, "per-page", "s", 50, "Per page")
	cmd.Flags().StringVar(&x.from, "from", "", "Filter from time (RFC3339)")
	cmd.Flags().StringVar(&x.to, "to", "", "Filter to time (RFC3339)")

	_ = cmd.RegisterFlagCompletionFunc("resource-type", completeOne(service.AuditResourceTypeNames()))
	_ = cmd.RegisterFlagCompletionFunc("action", completeOne(service.AuditActionNames()))
}

func (x *auditRecordsList) run(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("actor-id") {
		x.in.ActorID = &x.actorID
	}

	if x.from != "" {
		t, err := time.Parse(time.RFC3339, x.from)
		if err != nil {
			return service.BadRequestf("invalid --from value %q: expected RFC3339", x.from)
		}
		x.in.From = &t
	}

	if x.to != "" {
		t, err := time.Parse(time.RFC3339, x.to)
		if err != nil {
			return service.BadRequestf("invalid --to value %q: expected RFC3339", x.to)
		}
		x.in.To = &t
	}

	out, err := x.c.svc.AuditRecordsList(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Get
//

type auditRecordsGet struct {
	c  *Command
	in service.AuditRecordsGetInput
}

func (x *auditRecordsGet) flags(cmd *cobra.Command) {
	cmd.Use = "get ID"
	cmd.Args = cobra.ExactArgs(1)
}

func (x *auditRecordsGet) run(cmd *cobra.Command, args []string) error {
	i, err := parseIndex(args[0])
	if err != nil {
		return err
	}
	x.in.ID = i

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.AuditRecordsGet(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
