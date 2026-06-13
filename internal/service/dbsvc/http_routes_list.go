package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesList implements [service.Service].
func (s *Service) HTTPRoutesList(
	ctx context.Context,
	in service.HTTPRoutesListInput,
) (service.HTTPRoutesListOutput, error) {
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

	recs, err := s.db.HTTPRoutesGetByPayloadID(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	out := make([]service.HTTPRoute, len(recs))

	for i, r := range recs {
		out[i] = *httpRoute(*r, p.Subdomain)
	}

	return out, nil
}
