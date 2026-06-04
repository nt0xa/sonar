package dbsvc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesCreate implements [service.Service].
func (s *svc) HTTPRoutesCreate(
	ctx context.Context,
	in service.HTTPRoutesCreateInput,
) (*service.HTTPRoutesCreateOutput, error) {
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

	_, err = s.db.HTTPRoutesGetByPayloadMethodAndPath(ctx, database.HTTPRoutesGetByPayloadMethodAndPathParams{
		PayloadID: p.ID,
		Method:    string(in.Method),
		Path:      in.Path,
	})
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, fmt.Errorf("%w: http route for payload %q with method %q and path %q already exist",
			service.ErrConflict, in.PayloadName, in.Method, in.Path)
	}

	body, err := base64.StdEncoding.DecodeString(in.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: body: invalid base64 data", service.ErrValidation)
	}

	rec, err := s.db.HTTPRoutesCreate(ctx, database.HTTPRoutesCreateParams{
		PayloadID: p.ID,
		Method:    string(in.Method),
		Path:      in.Path,
		Code:      in.Code,
		Headers:   in.Headers,
		Body:      body,
		IsDynamic: in.IsDynamic,
	})
	if err != nil {
		return nil, err
	}

	return httpRoute(*rec, p.Subdomain), nil
}
