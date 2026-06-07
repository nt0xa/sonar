package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// DNSRecordsClear implements [service.Service].
func (s *svc) DNSRecordsClear(
	ctx context.Context,
	in service.DNSRecordsClearInput,
) (service.DNSRecordsClearOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	u := s.user(ctx)
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

	var recs []*database.DNSRecord

	if in.Name != "" {
		recs, err = s.db.DNSRecordsDeleteAllByPayloadIDAndName(ctx, p.ID, in.Name)
	} else {
		recs, err = s.db.DNSRecordsDeleteAllByPayloadID(ctx, p.ID)
	}
	if err != nil {
		return nil, err
	}

	out := make([]service.DNSRecord, len(recs))

	for i, r := range recs {
		out[i] = *dnsRecord(*r, p.Subdomain)
	}

	return out, nil
}
