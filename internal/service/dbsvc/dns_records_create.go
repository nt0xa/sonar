package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsCreate implements [service.Service].
func (s *Service) DNSRecordsCreate(
	ctx context.Context,
	in service.DNSRecordsCreateInput,
) (*service.DNSRecordsCreateOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	c, ok := service.CallerFrom(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, c.UserID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.PayloadName)
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
		return nil, service.Conflictf("dns records for payload %q with name %q and type %q already exist",
			in.PayloadName, in.Name, in.Type)
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

	return (*service.DNSRecordsCreateOutput)(dnsRecord(*rec, p.Subdomain)), nil
}
