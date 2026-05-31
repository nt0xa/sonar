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

// DNSRecordsClear implements [types.Service].
func (s *service) DNSRecordsClear(context.Context, types.DNSRecordsClearInput) (types.DNSRecordsClearOutput, error) {
	panic("unimplemented")
}

// DNSRecordsCreate implements [types.Service].
func (s *service) DNSRecordsCreate(context.Context, types.DNSRecordsCreateInput) (*types.DNSRecordsCreateOutput, error) {
	panic("unimplemented")
}

// DNSRecordsDelete implements [types.Service].
func (s *service) DNSRecordsDelete(context.Context, types.DNSRecordsDeleteInput) (*types.DNSRecordsDeleteOutput, error) {
	panic("unimplemented")
}

// DNSRecordsList implements [types.Service].
func (s *service) DNSRecordsList(context.Context, types.DNSRecordsListInput) (types.DNSRecordsListOutput, error) {
	panic("unimplemented")
}

// EventsGet implements [types.Service].
func (s *service) EventsGet(context.Context, types.EventsGetInput) (*types.EventsGetOutput, error) {
	panic("unimplemented")
}

// EventsList implements [types.Service].
func (s *service) EventsList(context.Context, types.EventsListInput) (types.EventsListOutput, error) {
	panic("unimplemented")
}

// HTTPRoutesClear implements [types.Service].
func (s *service) HTTPRoutesClear(context.Context, types.HTTPRoutesClearInput) (types.HTTPRoutesClearOutput, error) {
	panic("unimplemented")
}

// HTTPRoutesCreate implements [types.Service].
func (s *service) HTTPRoutesCreate(context.Context, types.HTTPRoutesCreateInput) (*types.HTTPRoutesCreateOutput, error) {
	panic("unimplemented")
}

// HTTPRoutesDelete implements [types.Service].
func (s *service) HTTPRoutesDelete(context.Context, types.HTTPRoutesDeleteInput) (*types.HTTPRoutesDeleteOutput, error) {
	panic("unimplemented")
}

// HTTPRoutesList implements [types.Service].
func (s *service) HTTPRoutesList(context.Context, types.HTTPRoutesListInput) (types.HTTPRoutesListOutput, error) {
	panic("unimplemented")
}

// HTTPRoutesUpdate implements [types.Service].
func (s *service) HTTPRoutesUpdate(context.Context, types.HTTPRoutesUpdateInput) (*types.HTTPRoutesUpdateOutput, error) {
	panic("unimplemented")
}

// PayloadsClear implements [types.Service].
func (s *service) PayloadsClear(context.Context, types.PayloadsClearInput) (types.PayloadsClearOutput, error) {
	panic("unimplemented")
}

// PayloadsDelete implements [types.Service].
func (s *service) PayloadsDelete(context.Context, types.PayloadsDeleteInput) (*types.PayloadsDeleteOutput, error) {
	panic("unimplemented")
}

// PayloadsList implements [types.Service].
func (s *service) PayloadsList(context.Context, types.PayloadsListInput) (types.PayloadsListOutput, error) {
	panic("unimplemented")
}

// ProfileGet implements [types.Service].
func (s *service) ProfileGet(context.Context, types.ProfileGetInput) (*types.ProfileGetOutput, error) {
	panic("unimplemented")
}

// UsersCreate implements [types.Service].
func (s *service) UsersCreate(context.Context, types.UsersCreateInput) (*types.UsersCreateOutput, error) {
	panic("unimplemented")
}

// UsersDelete implements [types.Service].
func (s *service) UsersDelete(context.Context, types.UsersDeleteInput) (*types.UsersDeleteOutput, error) {
	panic("unimplemented")
}

func New(db *database.DB, log *slog.Logger) types.Service {
	return &service{
		db:  db,
		log: log,
	}
}
