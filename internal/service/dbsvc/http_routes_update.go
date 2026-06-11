package dbsvc

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

// HTTPRoutesUpdate implements [service.Service].
func (s *Service) HTTPRoutesUpdate(
	ctx context.Context,
	in service.HTTPRoutesUpdateInput,
) (*service.HTTPRoutesUpdateOutput, error) {
	if p := in.Validate(); p != nil {
		return nil, service.Validation(p)
	}

	id, ok := service.GetUserID(ctx)
	if !ok {
		return nil, service.Unauthorized()
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, id, in.Payload)
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("payload with name %q not found", in.Payload)
	}
	if err != nil {
		return nil, err
	}

	rec, err := s.db.HTTPRoutesGetByPayloadIDAndIndex(ctx, p.ID, int(in.Index))
	if errors.Is(err, database.ErrNoRows) {
		return nil, service.NotFoundf("http route for payload %q with index %d not found",
			in.Payload, in.Index)
	}
	if err != nil {
		return nil, err
	}

	method := rec.Method
	if in.Method != nil {
		method = string(*in.Method)
	}

	path := rec.Path
	if in.Path != nil {
		path = *in.Path
	}

	code := rec.Code
	if in.Code != nil {
		code = *in.Code
	}

	headers := rec.Headers
	if in.Headers != nil {
		headers = in.Headers
	}

	body := rec.Body
	if in.Body != nil {
		body, err = base64.StdEncoding.DecodeString(*in.Body)
		if err != nil {
			return nil, service.Validation(map[string]string{"body": "invalid base64 data"})
		}
	}

	isDynamic := rec.IsDynamic
	if in.IsDynamic != nil {
		isDynamic = *in.IsDynamic
	}

	updated, err := s.db.HTTPRoutesUpdate(ctx, database.HTTPRoutesUpdateParams{
		ID:        rec.ID,
		PayloadID: rec.PayloadID,
		Method:    method,
		Path:      path,
		Code:      code,
		Headers:   headers,
		Body:      body,
		IsDynamic: isDynamic,
	})
	if err != nil {
		return nil, err
	}

	return httpRoute(*updated, p.Subdomain), nil
}
