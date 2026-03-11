package actionsdb

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
)

const (
	auditResourcePayload = "payload"
	auditResourceUser    = "user"
	auditResourceDNS     = "dns_record"
	auditResourceHTTP    = "http_route"

	auditOpCreate = "create"
	auditOpUpdate = "update"
	auditOpDelete = "delete"
	auditOpClear  = "clear"
)

type Auditable interface {
	AuditTarget() database.AuditTarget
	AuditData() database.AuditData
}

type auditableRecord struct {
	target database.AuditTarget
	data   database.AuditData
}

func (r auditableRecord) AuditTarget() database.AuditTarget {
	return r.target
}

func (r auditableRecord) AuditData() database.AuditData {
	if r.data == nil {
		return database.AuditData{}
	}
	return r.data
}

func newAuditable(target database.AuditTarget, data database.AuditData) Auditable {
	return auditableRecord{target: target, data: data}
}

func (act *dbactions) writeAudit(ctx context.Context, operation string, auditable Auditable) {
	if !act.audit || auditable == nil {
		return
	}

	u, err := GetUser(ctx)
	if err != nil || u == nil {
		act.log.Warn("failed to resolve actor for audit", "err", err)
		return
	}

	actor := database.AuditActor{
		ID:   &u.ID,
		Name: u.Name,
	}
	target := auditable.AuditTarget()
	data := auditable.AuditData()

	_, err = act.db.AuditRecordsCreate(ctx, database.AuditRecordsCreateParams{
		Operation: database.AuditRecordOperationType(operation),
		Actor:     actor,
		Target:    target,
		Data:      data,
	})
	if err != nil {
		act.log.Warn("failed to write audit record", "err", err, "resource", target.Type, "operation", operation)
	}
}
