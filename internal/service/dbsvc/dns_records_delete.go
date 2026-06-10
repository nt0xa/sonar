package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsDelete implements [service.Service].
func (s *svc) DNSRecordsDelete(
	ctx context.Context,
	in service.DNSRecordsDeleteInput,
) (*service.DNSRecordsDeleteOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := getUser(ctx)
	if u == nil {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.PayloadName)
	}
	if err != nil {
		return nil, err
	}

	rec, err := s.db.DNSRecordsGetByPayloadIDAndIndex(ctx, p.ID, int(in.Index))
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("dns record for payload %q with index %d not found",
			in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.DNSRecordsDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return dnsRecord(*rec, p.Subdomain), nil
}
