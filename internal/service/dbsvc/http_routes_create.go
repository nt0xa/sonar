package dbsvc

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesCreate implements [service.Service].
func (s *svc) HTTPRoutesCreate(
	ctx context.Context,
	in service.HTTPRoutesCreateInput,
) (*service.HTTPRoutesCreateOutput, error) {
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

	_, err = s.db.HTTPRoutesGetByPayloadMethodAndPath(ctx, database.HTTPRoutesGetByPayloadMethodAndPathParams{
		PayloadID: p.ID,
		Method:    string(in.Method),
		Path:      in.Path,
	})
	if err != nil && !errors.Is(err, database.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		return nil, service.Conflictf("http route for payload %q with method %q and path %q already exist",
			in.PayloadName, in.Method, in.Path)
	}

	body, err := base64.StdEncoding.DecodeString(in.Body)
	if err != nil {
		return nil, service.Validation(map[string]string{"body": "invalid base64 data"})
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
