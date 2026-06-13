package dbsvc

import (
	"context"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesDelete implements [service.Service].
func (s *Service) HTTPRoutesDelete(
	ctx context.Context,
	in service.HTTPRoutesDeleteInput,
) (*service.HTTPRoutesDeleteOutput, error) {
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

	rec, err := s.db.HTTPRoutesGetByPayloadIDAndIndex(ctx, p.ID, int(in.Index))
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("http route for payload %q with index %d not found",
			in.PayloadName, in.Index)
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.HTTPRoutesDelete(ctx, rec.ID); err != nil {
		return nil, err
	}

	return httpRoute(*rec, p.Subdomain), nil
}
