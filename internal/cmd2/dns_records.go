package cmd2

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) dnsRecordsCreate(cmd *cobra.Command) cmdx.RunFunc {
	var in service.DNSRecordsCreateInput

	cmd.Use = "new VALUES..."
	cmd.Args = cobra.MinimumNArgs(1)

	in.Type = service.DNSRecordTypeA
	in.Strategy = service.DNSRecordStrategyAll

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Subdomain")
	cmd.Flags().IntVarP(&in.TTL, "ttl", "l", 60, "Record TTL (in seconds)")
	cmd.Flags().VarP(&in.Type, "type", "t",
		fmt.Sprintf("Record type (one of %s)", strings.Join(service.DNSRecordTypeNames(), ", ")))
	cmd.Flags().VarP(&in.Strategy, "strategy", "s",
		fmt.Sprintf("Strategy for multiple records (one of %s)", strings.Join(service.DNSRecordStrategyNames(), ", ")))

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("type", completeOne(service.DNSRecordTypeNames()))
	_ = cmd.RegisterFlagCompletionFunc("strategy", completeOne(service.DNSRecordStrategyNames()))

	return func(cmd *cobra.Command, args []string) error {
		in.Values = args

		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.DNSRecordsCreate(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) dnsRecordsDelete(cmd *cobra.Command) cmdx.RunFunc {
	var in service.DNSRecordsDeleteInput

	cmd.Use = "del INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = c.completeDNSRecord

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

		out, err := c.svc.DNSRecordsDelete(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) dnsRecordsList(cmd *cobra.Command) cmdx.RunFunc {
	var in service.DNSRecordsListInput

	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.DNSRecordsList(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}

func (c *Command) dnsRecordsClear(cmd *cobra.Command) cmdx.RunFunc {
	var in service.DNSRecordsClearInput

	cmd.Use = "clr"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&in.Name, "name", "n", "", "Subdomain")

	_ = cmd.RegisterFlagCompletionFunc("payload", c.completePayloadName)

	return func(cmd *cobra.Command, args []string) error {
		if err := validate(in); err != nil {
			return err
		}

		out, err := c.svc.DNSRecordsClear(cmd.Context(), in)
		if err != nil {
			return err
		}

		return setResult(cmd.Context(), out)
	}
}
