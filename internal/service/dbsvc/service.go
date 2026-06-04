package dbsvc

import (
	"context"
	"log/slog"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

type svc struct {
	db  *database.DB
	log *slog.Logger
}

// AuditRecordsGet implements [service.Service].
func (s *svc) AuditRecordsGet(context.Context, service.AuditRecordsGetInput) (*service.AuditRecordsGetOutput, error) {
	panic("unimplemented")
}

// AuditRecordsList implements [service.Service].
func (s *svc) AuditRecordsList(context.Context, service.AuditRecordsListInput) (service.AuditRecordsListOutput, error) {
	panic("unimplemented")
}

func New(db *database.DB, log *slog.Logger) service.Service {
	return &svc{
		db:  db,
		log: log,
	}
}
