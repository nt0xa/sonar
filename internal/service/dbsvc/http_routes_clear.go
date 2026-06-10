package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesClear implements [service.Service].
func (s *svc) HTTPRoutesClear(
	ctx context.Context,
	in service.HTTPRoutesClearInput,
) (service.HTTPRoutesClearOutput, error) {
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

	var recs []*database.HTTPRoute

	if in.Path != "" {
		recs, err = s.db.HTTPRoutesDeleteAllByPayloadIDAndPath(ctx, p.ID, in.Path)
	} else {
		recs, err = s.db.HTTPRoutesDeleteAllByPayloadID(ctx, p.ID)
	}
	if err != nil {
		return nil, err
	}

	out := make([]service.HTTPRoute, len(recs))

	for i, r := range recs {
		out[i] = *httpRoute(*r, p.Subdomain)
	}

	return out, nil
}
