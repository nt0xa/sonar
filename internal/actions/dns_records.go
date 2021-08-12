package actions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/valid"
)

type DNSActions interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateParams) (DNSRecordsCreateResult, errors.Error)
	DNSRecordsDelete(context.Context, DNSRecordsDeleteParams) (DNSRecordsDeleteResult, errors.Error)
	DNSRecordsList(context.Context, DNSRecordsListParams) (DNSRecordsListResult, errors.Error)
}

type DNSRecordsHandler interface {
	DNSRecordsCreate(context.Context, DNSRecordsCreateResult)
	DNSRecordsList(context.Context, DNSRecordsListResult)
	DNSRecordsDelete(context.Context, DNSRecordsDeleteResult)
}

type DNSRecord struct {
	Index            int64     `json:"index"`
	PayloadSubdomain string    `json:"payloadSubdomain"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	TTL              int       `json:"ttl"`
	Values           []string  `json:"values"`
	Strategy         string    `json:"strategy"`
	CreatedAt        time.Time `json:"createdAt"`
}

//
// Create
//

type DNSRecordsCreateParams struct {
	PayloadName string   `err:"payloadName" json:"payloadName"`
	Name        string   `err:"name"        json:"name"`
	TTL         int      `err:"ttl"         json:"ttl"`
	Type        string   `err:"type"        json:"type"`
	Values      []string `err:"values"      json:"values"`
	Strategy    string   `err:"strategy"    json:"strategy"`
}

func (p DNSRecordsCreateParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Name, validation.Required, validation.By(valid.Subdomain)),
		validation.Field(&p.Type, valid.OneOf(models.DNSTypesAll, false)),
		validation.Field(&p.Values, validation.Required, validation.Each(valid.DNSRecord(p.Type))),
		validation.Field(&p.Strategy, valid.OneOf(models.DNSStrategiesAll, true)),
	)
}

type DNSRecordsCreateResult *DNSRecord

func DNSRecordsCreateCommand(p *DNSRecordsCreateParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "new VALUES",
		Short: "Create new DNS records",
		Args:  atLeastOneArg("VALUES"),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")
	cmd.Flags().StringVarP(&p.Name, "name", "n", "", "Subdomain")
	cmd.Flags().IntVarP(&p.TTL, "ttl", "l", 60, "Record TTL (in seconds)")
	cmd.Flags().StringVarP(&p.Type, "type", "t", "A",
		fmt.Sprintf("Record type (one of %s)", quoteAndJoin(models.DNSTypesAll)))
	cmd.Flags().StringVarP(&p.Strategy, "strategy", "s", models.DNSStrategyAll,
		fmt.Sprintf("Strategy for multiple records (one of %s)", quoteAndJoin(models.DNSStrategiesAll)))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		p.Values = args
		return nil
	}
}

//
// Delete
//

type DNSRecordsDeleteParams struct {
	PayloadName string `err:"payload" path:"payload"`
	Index       int64  `err:"index"   path:"index"`
}

func (p DNSRecordsDeleteParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
		validation.Field(&p.Index, validation.Required),
	)
}

type DNSRecordsDeleteResult *DNSRecord

func DNSRecordsDeleteCommand(p *DNSRecordsDeleteParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "del INDEX",
		Short: "Delete DNS record",
		Long:  "Delete DNS record identified by INDEX",
		Args:  oneArg("INDEX"),
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.Index = i
		return nil
	}
}

//
// List
//

type DNSRecordsListParams struct {
	PayloadName string `err:"payload" path:"payload"`
}

func (p DNSRecordsListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.PayloadName, validation.Required),
	)
}

type DNSRecordsListResult []*DNSRecord

func DNSRecordsListCommand(p *DNSRecordsListParams) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records",
	}

	cmd.Flags().StringVarP(&p.PayloadName, "payload", "p", "", "Payload name")

	return cmd, nil
}
