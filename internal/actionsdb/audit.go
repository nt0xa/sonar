package actionsdb

import (
	"context"
	"reflect"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils"
)

func (act *dbactions) auditCreate(ctx context.Context, rec any) {
	act.writeAudit(ctx, database.AuditRecordActionTypeCreate, rec)
}

func (act *dbactions) auditUpdate(ctx context.Context, rec any) {
	act.writeAudit(ctx, database.AuditRecordActionTypeUpdate, rec)
}

func (act *dbactions) auditDelete(ctx context.Context, rec any) {
	act.writeAudit(ctx, database.AuditRecordActionTypeDelete, rec)
}

func (act *dbactions) writeAudit(ctx context.Context, action database.AuditRecordActionType, rec any) {
	if !act.audit || rec == nil {
		return
	}

	u, err := GetUser(ctx)
	if err != nil || u == nil {
		act.log.Warn("failed to resolve actor for audit", "err", err)
		return
	}

	resourceType, ok := auditResourceTypeByRecord(rec)
	if !ok {
		act.log.Warn("skip audit: unsupported record type", "type", reflect.TypeOf(rec))
		return
	}

	source := database.AuditRecordSourceTypeAPI
	if s, err := GetSource(ctx); err == nil {
		switch database.AuditRecordSourceType(s) {
		case database.AuditRecordSourceTypeAPI,
			database.AuditRecordSourceTypeTelegram,
			database.AuditRecordSourceTypeLark,
			database.AuditRecordSourceTypeSlack:
			source = database.AuditRecordSourceType(s)
		}
	}

	actorMeta := database.AuditActorMetadata{}
	switch source {
	case database.AuditRecordSourceTypeTelegram:
		if u.TelegramID != nil {
			actorMeta["telegram_id"] = *u.TelegramID
		}
	case database.AuditRecordSourceTypeLark:
		if u.LarkID != nil {
			actorMeta["lark_id"] = *u.LarkID
		}
	case database.AuditRecordSourceTypeSlack:
		if u.SlackID != nil {
			actorMeta["slack_id"] = *u.SlackID
		}
	}

	resource := utils.StructToMap(rec)

	_, err = act.db.AuditRecordsCreate(ctx, database.AuditRecordsCreateParams{
		Action:        action,
		ResourceType:  resourceType,
		Source:        source,
		ActorID:       &u.ID,
		ActorName:     u.Name,
		ActorMetadata: actorMeta,
		Resource:      resource,
	})
	if err != nil {
		act.log.Warn("failed to write audit record", "err", err, "resourceType", resourceType, "action", action)
	}
}

func auditResourceTypeByRecord(rec any) (database.AuditRecordResourceType, bool) {
	switch rec.(type) {
	case actions.Payload, *actions.Payload:
		return database.AuditRecordResourceTypePayload, true
	case actions.DNSRecord, *actions.DNSRecord:
		return database.AuditRecordResourceTypeDNSRecord, true
	case actions.HTTPRoute, *actions.HTTPRoute:
		return database.AuditRecordResourceTypeHTTPRoute, true
	case actions.User, *actions.User:
		return database.AuditRecordResourceTypeUser, true
	default:
		return database.AuditRecordResourceType(""), false
	}
}
