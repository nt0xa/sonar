package cmd2

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/cmdx"
)

func (c *Command) addDNS(g *cmdx.Command) {
	create := &dnsRecordsCreate{c: c}
	g.Add("new", "Create new DNS records", create.run, create.flags)

	del := &dnsRecordsDelete{c: c}
	g.Add("del", "Delete DNS record", del.run, del.flags)

	list := &dnsRecordsList{c: c}
	g.Add("list", "List DNS records", list.run, list.flags)

	clear := &dnsRecordsClear{c: c}
	g.Add("clr", "Delete multiple DNS records", clear.run, clear.flags)
}

//
// Create
//

type dnsRecordsCreate struct {
	c  *Command
	in service.DNSRecordsCreateInput
}

func (x *dnsRecordsCreate) flags(cmd *cobra.Command) {
	cmd.Use = "new VALUES..."
	cmd.Args = cobra.MinimumNArgs(1)

	x.in.Type = service.DNSRecordTypeA
	x.in.Strategy = service.DNSRecordStrategyAll

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&x.in.Name, "name", "n", "", "Subdomain")
	cmd.Flags().IntVarP(&x.in.TTL, "ttl", "l", 60, "Record TTL (in seconds)")
	cmd.Flags().VarP(dnsTypeValue{&x.in.Type}, "type", "t",
		fmt.Sprintf("Record type (one of %s)", strings.Join(service.DNSRecordTypeNames(), ", ")))
	cmd.Flags().VarP(dnsStrategyValue{&x.in.Strategy}, "strategy", "s",
		fmt.Sprintf("Strategy for multiple records (one of %s)", strings.Join(service.DNSRecordStrategyNames(), ", ")))

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
	_ = cmd.RegisterFlagCompletionFunc("type", completeOne(service.DNSRecordTypeNames()))
	_ = cmd.RegisterFlagCompletionFunc("strategy", completeOne(service.DNSRecordStrategyNames()))
}

func (x *dnsRecordsCreate) run(cmd *cobra.Command, args []string) error {
	x.in.Values = args

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.DNSRecordsCreate(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Delete
//

type dnsRecordsDelete struct {
	c  *Command
	in service.DNSRecordsDeleteInput
}

func (x *dnsRecordsDelete) flags(cmd *cobra.Command) {
	cmd.Use = "del INDEX"
	cmd.Args = cobra.ExactArgs(1)
	cmd.ValidArgsFunction = x.c.completeDNSRecord

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *dnsRecordsDelete) run(cmd *cobra.Command, args []string) error {
	i, err := parseIndex(args[0])
	if err != nil {
		return err
	}
	x.in.Index = i

	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.DNSRecordsDelete(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// List
//

type dnsRecordsList struct {
	c  *Command
	in service.DNSRecordsListInput
}

func (x *dnsRecordsList) flags(cmd *cobra.Command) {
	cmd.Use = "list"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *dnsRecordsList) run(cmd *cobra.Command, args []string) error {
	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.DNSRecordsList(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}

//
// Clear
//

type dnsRecordsClear struct {
	c  *Command
	in service.DNSRecordsClearInput
}

func (x *dnsRecordsClear) flags(cmd *cobra.Command) {
	cmd.Use = "clr"
	cmd.Args = cobra.NoArgs

	cmd.Flags().StringVarP(&x.in.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&x.in.Name, "name", "n", "", "Subdomain")

	_ = cmd.RegisterFlagCompletionFunc("payload", x.c.completePayloadName)
}

func (x *dnsRecordsClear) run(cmd *cobra.Command, args []string) error {
	if err := validate(x.in); err != nil {
		return err
	}

	out, err := x.c.svc.DNSRecordsClear(cmd.Context(), x.in)
	if err != nil {
		return err
	}

	return setResult(cmd.Context(), out)
}
