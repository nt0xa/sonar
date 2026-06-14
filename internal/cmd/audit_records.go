package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
)

func (c *Command) auditRecordsList(cmd *cobra.Command) runFunc {
	var (
		in      service.AuditRecordsListInput
		actorID int64
		from    string
		to      string
	)

	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().Int64Var(&actorID, "actor-id", 0, "Filter by actor ID")
	cmd.Flags().StringVar(&in.ActorName, "actor-name", "", "Filter by actor name")
	cmd.Flags().Var(&in.ResourceType, "resource-type", "Filter by resource type")
	cmd.Flags().Var(&in.Action, "action", "Filter by action")
	cmd.Flags().UintVarP(&in.Page, "page", "p", 1, "Page")
	cmd.Flags().UintVarP(&in.PerPage, "per-page", "s", 50, "Per page")
	cmd.Flags().StringVar(&from, "from", "", "Filter from time (RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "Filter to time (RFC3339)")

	_ = cmd.RegisterFlagCompletionFunc("resource-type", completeOne(service.AuditResourceTypeNames()))
	_ = cmd.RegisterFlagCompletionFunc("action", completeOne(service.AuditActionNames()))

	return func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("actor-id") {
			in.ActorID = &actorID
		}

		if from != "" {
			t, err := time.Parse(time.RFC3339, from)
			if err != nil {
				return service.BadRequestf("invalid --from value %q: expected RFC3339", from)
			}
			in.From = &t
		}

		if to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				return service.BadRequestf("invalid --to value %q: expected RFC3339", to)
			}
			in.To = &t
		}

		out, err := c.svc.AuditRecordsList(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) auditRecordsGet(cmd *cobra.Command) runFunc {
	var in service.AuditRecordsGetInput

	cmd.Use = "get ID"
	cmd.Args = cobra.ExactArgs(1)

	return func(cmd *cobra.Command, args []string) error {
		i, err := parseIndex(args[0])
		if err != nil {
			return err
		}
		in.ID = i

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.AuditRecordsGet(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
