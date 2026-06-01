package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

// HTTPRoutesUpdate implements [types.Service].
func (s *service) HTTPRoutesUpdate(
	ctx context.Context,
	in types.HTTPRoutesUpdateInput,
) (*types.HTTPRoutesUpdateOutput, error) {
	u := s.user(ctx)
	if u == nil {
		return nil, types.ErrUnauthorized
	}

	p, err := s.db.PayloadsGetByUserAndName(ctx, u.ID, in.Payload)
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: payload with name %q not found", types.ErrNotFound, in.Payload)
	}
	if err != nil {
		return nil, err
	}

	rec, err := s.db.HTTPRoutesGetByPayloadIDAndIndex(ctx, p.ID, int(in.Index))
	if errors.Is(err, database.ErrNoRows) {
		return nil, fmt.Errorf("%w: http route for payload %q with index %d not found",
			types.ErrNotFound, in.Payload, in.Index)
	}
	if err != nil {
		return nil, err
	}

	method := rec.Method
	if in.Method != nil {
		method = *in.Method
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
			return nil, fmt.Errorf("%w: body: invalid base64 data", types.ErrValidation)
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
