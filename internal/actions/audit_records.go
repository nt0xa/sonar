package actions

import (
	"context"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
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
	auditActions       = []string{"create", "update", "delete"}
)

type AuditRecordsActions interface {
	AuditRecordsList(context.Context, AuditRecordsListParams) (AuditRecordsListResult, errors.Error)
	AuditRecordsGet(context.Context, AuditRecordsGetParams) (*AuditRecordsGetResult, errors.Error)
}

type AuditRecord struct {
	ID           int64          `json:"id"`
	UUID         uuid.UUID      `json:"uuid"`
	CreatedAt    time.Time      `json:"createdAt"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resourceType"`
	Source       string         `json:"source"`
	ActorID      *int64         `json:"actorId,omitempty"`
	ActorName    string         `json:"actorName,omitempty"`
	ActorMeta    map[string]any `json:"actorMetadata"`
	Resource     map[string]any `json:"resource"`
}

//
// List
//

type AuditRecordsListParams struct {
	ActorID      *int64     `query:"actorId,omitempty"`
	ActorName    string     `query:"actorName,omitempty"`
	ResourceType string     `query:"resourceType,omitempty"`
	Action       string     `query:"action,omitempty"`
	From         *time.Time `query:"from,omitempty"`
	To           *time.Time `query:"to,omitempty"`
	Page         uint       `query:"page,omitempty"`
	PerPage      uint       `query:"perPage,omitempty"`
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
		actorID int64
	)

	cmd.Flags().Int64Var(&actorID, "actor-id", 0, "Filter by actor ID")
	cmd.Flags().StringVar(&p.ActorName, "actor-name", "", "Filter by actor name")
	cmd.Flags().StringVar(&p.ResourceType, "resource-type", "", "Filter by resource type")
	cmd.Flags().StringVar(&p.Action, "action", "", "Filter by action")
	cmd.Flags().UintVarP(&p.Page, "page", "p", 1, "Page")
	cmd.Flags().UintVarP(&p.PerPage, "per-page", "s", 50, "Per page")

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
