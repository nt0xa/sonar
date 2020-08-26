package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *command) DNSRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
	}

	cmd.AddCommand(c.DNSRecordsCreate())
	cmd.AddCommand(c.DNSRecordsDelete())
	cmd.AddCommand(c.DNSRecordsList())

	return cmd
}

func (c *command) DNSRecordsCreate() *cobra.Command {
	var p actions.DNSRecordsCreateParams

	cmd := &cobra.Command{
		Use:   "new VALUES",
		Short: "Create new DNS records",
		Long:  "Create new DNS records",
		Args:  AtLeastOneArg("VALUES"),
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Values = args

			res, err := c.actions.DNSRecordsCreate(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.DNSRecordsCreate(cmd.Context(), res)

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

func (c *command) DNSRecordsDelete() *cobra.Command {
	var p actions.DNSRecordsDeleteParams

	cmd := &cobra.Command{
		Use:   "del",
		Short: "Delete DNS records",
		Long:  "Delete DNS records",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			p.Type = strings.ToUpper(p.Type)

			res, err := c.actions.DNSRecordsDelete(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.DNSRecordsDelete(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Name, "name", "n", "", "Subdomain")
	cmd.Flags().StringVarP(&p.Type, "type", "t", "A",
		fmt.Sprintf("Record type (one of %s)", quoteAndJoin(models.DNSTypesAll)))

	return cmd
}

func (c *command) DNSRecordsList() *cobra.Command {
	var p actions.DNSRecordsListParams

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records for payload",
		Long:  "List DNS records for payload with name PAYLOAD",
		RunE: RunE(func(cmd *cobra.Command, args []string) errors.Error {
			res, err := c.actions.DNSRecordsList(cmd.Context(), p)
			if err != nil {
				return err
			}

			c.handler.DNSRecordsList(cmd.Context(), res)

			return nil
		}),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd
}
