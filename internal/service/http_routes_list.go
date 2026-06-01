package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// HTTPRoutesList implements [types.Service].
func (s *service) HTTPRoutesList(
	ctx context.Context,
	in types.HTTPRoutesListInput,
) (types.HTTPRoutesListOutput, error) {
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

	recs, err := s.db.HTTPRoutesGetByPayloadID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	out := make([]types.HTTPRoute, len(recs))

	for i, r := range recs {
		out[i] = *httpRoute(*r, p.Subdomain)
	}

	return out, nil
}
