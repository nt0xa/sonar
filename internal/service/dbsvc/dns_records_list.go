package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsList implements [service.Service].
func (s *Service) DNSRecordsList(
	ctx context.Context,
	in service.DNSRecordsListInput,
) (service.DNSRecordsListOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	id, ok := service.GetUserID(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, id, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.PayloadName)
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
