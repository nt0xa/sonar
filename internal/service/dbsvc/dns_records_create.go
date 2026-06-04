package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsCreate implements [service.Service].
func (s *svc) DNSRecordsCreate(
	ctx context.Context,
	in service.DNSRecordsCreateInput,
) (*service.DNSRecordsCreateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, service.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", service.ErrNotFound, in.PayloadName)
	}
	if err != nil {
		return nil, err
	}

	_, err = s.db.DNSRecordsGetByPayloadNameAndType(ctx, database.DNSRecordsGetByPayloadNameAndTypeParams{
		PayloadID: p.ID,
		Name:      in.Name,
		Type:      database.DNSRecordType(in.Type),
	})
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: dns records for payload %q with name %q and type %q already exist",
			service.ErrConflict, in.PayloadName, in.Name, in.Type)
	}

	rec, err := s.db.DNSRecordsCreate(ctx, database.DNSRecordsCreateParams{
		PayloadID: p.ID,
		Name:      in.Name,
		TTL:       in.TTL,
		Type:      database.DNSRecordType(in.Type),
		Values:    in.Values,
		Strategy:  database.DNSStrategy(in.Strategy),
	})
	if err != nil {
		return nil, err
	}

	return dnsRecord(*rec, p.Subdomain), nil
}
