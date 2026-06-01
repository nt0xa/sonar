package service

import (
	"context"
	"log/slog"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

type service struct {
	db  *database.DB
	log *slog.Logger
}

// AuditRecordsGet implements [types.Service].
func (s *service) AuditRecordsGet(context.Context, types.AuditRecordsGetInput) (*types.AuditRecordsGetOutput, error) {
	panic("unimplemented")
}

// AuditRecordsList implements [types.Service].
func (s *service) AuditRecordsList(context.Context, types.AuditRecordsListInput) (types.AuditRecordsListOutput, error) {
	panic("unimplemented")
}

func New(db *database.DB, log *slog.Logger) types.Service {
	return &service{
		db:  db,
		log: log,
	}
}
