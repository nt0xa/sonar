package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// DNSRecordsCreate implements [types.Service].
func (s *service) DNSRecordsCreate(
	ctx context.Context,
	in types.DNSRecordsCreateInput,
) (*types.DNSRecordsCreateOutput, error) {
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

	_, err = s.db.DNSRecordsGetByPayloadNameAndType(ctx, database.DNSRecordsGetByPayloadNameAndTypeParams{
		PayloadID: p.ID,
		Name:      in.Name,
		Type:      database.DNSRecordType(strings.ToUpper(in.Type)),
	})
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: dns records for payload %q with name %q and type %q already exist",
			types.ErrConflict, in.PayloadName, in.Name, in.Type)
	}

	rec, err := s.db.DNSRecordsCreate(ctx, database.DNSRecordsCreateParams{
		PayloadID: p.ID,
		Name:      in.Name,
		TTL:       in.TTL,
		Type:      database.DNSRecordType(strings.ToUpper(in.Type)),
		Values:    in.Values,
		Strategy:  database.DNSStrategy(in.Strategy),
	})
	if err != nil {
		return nil, err
	}

	return dnsRecord(*rec, p.Subdomain), nil
}
