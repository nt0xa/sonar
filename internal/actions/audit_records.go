package actions

import (
	"context"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/cobra"

	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/internal/utils/valid"
)

const (
	AuditRecordsListResultID = "audit_records/list"
	AuditRecordsGetResultID  = "audit_records/get"
)

var (
	auditResourceTypes = []string{"payload", "user", "dns_record", "http_route"}
	auditActions       = []string{"create", "update", "delete", "clear"}
)

type AuditRecordsActions interface {
	AuditRecordsList(context.Context, AuditRecordsListParams) (AuditRecordsListResult, errors.Error)
	AuditRecordsGet(context.Context, AuditRecordsGetParams) (*AuditRecordsGetResult, errors.Error)
}

type AuditRecord struct {
	ID           int64          `json:"id"`
	ActorID      *int64         `json:"actorId,omitempty"`
	ActorName    string         `json:"actorName,omitempty"`
	ResourceType string         `json:"resourceType"`
	ResourceID   *int64         `json:"resourceId,omitempty"`
	ResourceKey  string         `json:"resourceKey"`
	Action       string         `json:"action"`
	PayloadID    *int64         `json:"payloadId,omitempty"`
	PayloadName  string         `json:"payloadName,omitempty"`
	Meta         map[string]any `json:"meta"`
	CreatedAt    time.Time      `json:"createdAt"`
}

//
// List
//

type AuditRecordsListParams struct {
	ActorID      *int64     `query:"actorId"`
	ActorName    string     `query:"actorName"`
	ResourceType string     `query:"resourceType"`
	ResourceID   *int64     `query:"resourceId"`
	ResourceKey  string     `query:"resourceKey"`
	Action       string     `query:"action"`
	PayloadID    *int64     `query:"payloadId"`
	PayloadName  string     `query:"payloadName"`
	From         *time.Time `query:"from"`
	To           *time.Time `query:"to"`
	Limit        uint       `query:"limit"`
	Offset       uint       `query:"offset"`
}

func (p AuditRecordsListParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.ResourceType, validation.When(p.ResourceType != "",
			valid.OneOf(auditResourceTypes, false))),
		validation.Field(&p.Action, validation.When(p.Action != "",
			valid.OneOf(auditActions, false))),
		validation.Field(&p.From),
		validation.Field(&p.To),
	)
}

type AuditRecordsListResult []AuditRecord

func (r AuditRecordsListResult) ResultID() string {
	return AuditRecordsListResultID
}

func AuditRecordsListCommand(acts *Actions, p *AuditRecordsListParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit records",
	}

	var (
		actorID    int64
		resourceID int64
		payloadID  int64
	)

	cmd.Flags().Int64Var(&actorID, "actor-id", 0, "Filter by actor ID")
	cmd.Flags().StringVar(&p.ActorName, "actor-name", "", "Filter by actor name")
	cmd.Flags().StringVar(&p.ResourceType, "resource-type", "", "Filter by resource type")
	cmd.Flags().Int64Var(&resourceID, "resource-id", 0, "Filter by resource ID")
	cmd.Flags().StringVar(&p.ResourceKey, "resource-key", "", "Filter by resource key")
	cmd.Flags().StringVar(&p.Action, "action", "", "Filter by action")
	cmd.Flags().Int64Var(&payloadID, "payload-id", 0, "Filter by payload ID")
	cmd.Flags().StringVar(&p.PayloadName, "payload-name", "", "Filter by payload name")
	cmd.Flags().UintVarP(&p.Limit, "limit", "l", 50, "Limit")
	cmd.Flags().UintVarP(&p.Offset, "offset", "o", 0, "Offset")

	var from string
	var to string
	cmd.Flags().StringVar(&from, "from", "", "Filter from time (RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "Filter to time (RFC3339)")

	_ = cmd.RegisterFlagCompletionFunc("resource-type", completeOne(auditResourceTypes))
	_ = cmd.RegisterFlagCompletionFunc("action", completeOne(auditActions))

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		if cmd.Flags().Changed("actor-id") {
			p.ActorID = &actorID
		} else {
			p.ActorID = nil
		}

		if cmd.Flags().Changed("resource-id") {
			p.ResourceID = &resourceID
		} else {
			p.ResourceID = nil
		}

		if cmd.Flags().Changed("payload-id") {
			p.PayloadID = &payloadID
		} else {
			p.PayloadID = nil
		}

		if from != "" {
			t, err := time.Parse(time.RFC3339, from)
			if err != nil {
				return errors.Validationf("invalid --from value %q: expected RFC3339", from)
			}
			p.From = &t
		}

		if to != "" {
			t, err := time.Parse(time.RFC3339, to)
			if err != nil {
				return errors.Validationf("invalid --to value %q: expected RFC3339", to)
			}
			p.To = &t
		}

		return nil
	}
}

//
// Get
//

type AuditRecordsGetParams struct {
	ID int64 `path:"id"`
}

func (p AuditRecordsGetParams) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.ID, validation.Required),
	)
}

type AuditRecordsGetResult struct {
	AuditRecord
}

func (r AuditRecordsGetResult) ResultID() string {
	return AuditRecordsGetResultID
}

func AuditRecordsGetCommand(acts *Actions, p *AuditRecordsGetParams, local bool) (*cobra.Command, PrepareCommandFunc) {
	cmd := &cobra.Command{
		Use:   "get ID",
		Short: "Get audit record by ID",
		Args:  oneArg("ID"),
	}

	return cmd, func(cmd *cobra.Command, args []string) errors.Error {
		i, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Validationf("invalid integer value %q", args[0])
		}
		p.ID = i
		return nil
	}
}
