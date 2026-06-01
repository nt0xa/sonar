package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// DNSRecordsDelete implements [types.Service].
func (s *service) DNSRecordsDelete(
	ctx context.Context,
	in types.DNSRecordsDeleteInput,
) (*types.DNSRecordsDeleteOutput, error) {
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

	rec, err := s.db.DNSRecordsGetByPayloadIDAndIndex(ctx, p.ID, int(in.Index))
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: dns record for payload %q with index %d not found",
			types.ErrNotFound, in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.DNSRecordsDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return dnsRecord(*rec, p.Subdomain), nil
}
