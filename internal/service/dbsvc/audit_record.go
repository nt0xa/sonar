package dbsvc

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

func auditRecord(m database.AuditRecord) *service.AuditRecord {
	return &service.AuditRecord{
		ID:           m.ID,
		UUID:         m.UUID.String(),
		CreatedAt:    m.CreatedAt,
		Action:       service.AuditAction(m.Action),
		ResourceType: service.AuditResourceType(m.ResourceType),
		Source:       service.AuditSource(m.Source),
		ActorID:      m.ActorID,
		ActorName:    m.ActorName,
		ActorMeta:    m.ActorMetadata,
		Resource:     m.Resource,
	}
}
