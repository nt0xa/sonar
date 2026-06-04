package dbsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsList implements [service.Service].
func (s *svc) DNSRecordsList(
	ctx context.Context,
	in service.DNSRecordsListInput,
) (service.DNSRecordsListOutput, error) {
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

	recs, err := s.db.DNSRecordsGetByPayloadID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	out := make([]service.DNSRecord, len(recs))

	for i, r := range recs {
		out[i] = *dnsRecord(*r, p.Subdomain)
	}

	return out, nil
}
