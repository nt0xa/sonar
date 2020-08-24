package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Command) DNS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}

	cmd.AddCommand(c.CreateDNSRecord())
	cmd.AddCommand(c.DeleteDNSRecord())
	cmd.AddCommand(c.ListDNSRecords())

	return cmd
}

func (c *Command) CreateDNSRecord() *cobra.Command {
	var p actions.CreateDNSRecordParams

	cmd := &cobra.Command{
		Use:   "new VALUES",
		Short: "Create new DNS records",
		Long:  "Create new DNS records",
		Args:  AtLeastOneArg("VALUES"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Values = args
			p.Type = strings.ToUpper(p.Type)

			res, err := c.Actions.CreateDNSRecord(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Name, "name", "n", "", "Subdomain")
	cmd.Flags().IntVarP(&p.TTL, "ttl", "l", 60, "Record TTL (in seconds)")
	cmd.Flags().StringVarP(&p.Type, "type", "t", "A",
		fmt.Sprintf("Record type (one of %s)", quoteAndJoin(models.DNSTypesAll)))
	cmd.Flags().StringVarP(&p.Strategy, "strategy", "s", models.DNSStrategyAll,
		fmt.Sprintf("Strategy for multiple records (one of %s)", quoteAndJoin(models.DNSStrategiesAll)))

	return cmd
}

func (c *Command) DeleteDNSRecord() *cobra.Command {
	var p actions.DeleteDNSRecordParams

	cmd := &cobra.Command{
		Use:   "del",
		Short: "Delete DNS records",
		Long:  "Delete DNS records",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Type = strings.ToUpper(p.Type)

			res, err := c.Actions.DeleteDNSRecord(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Name, "name", "n", "", "Subdomain")
	cmd.Flags().StringVarP(&p.Type, "type", "t", "A",
		fmt.Sprintf("Record type (one of %s)", quoteAndJoin(models.DNSTypesAll)))

	return cmd
}

func (c *Command) ListDNSRecords() *cobra.Command {
	var p actions.ListDNSRecordsParams

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records for payload",
		Long:  "List DNS records for payload with name PAYLOAD",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			res, err := c.Actions.ListDNSRecords(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.ResultHandler(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd
}
