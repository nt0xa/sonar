package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// DNSRecordsClear implements [types.Service].
func (s *service) DNSRecordsClear(
	ctx context.Context,
	in types.DNSRecordsClearInput,
) (types.DNSRecordsClearOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.PayloadName)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", types.ErrNotFound, in.PayloadName)
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

	out := make([]types.DNSRecord, len(recs))

	for i, r := range recs {
		out[i] = *dnsRecord(*r, p.Subdomain)
	}

	return out, nil
}
