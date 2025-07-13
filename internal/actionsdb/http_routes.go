package actionsdb

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strings"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func HTTPRoute(m models.HTTPRoute, payloadSubdomain string) actions.HTTPRoute {
	return actions.HTTPRoute{
		Index:            m.Index,
		PayloadSubdomain: payloadSubdomain,
		Method:           m.Method,
		Path:             m.Path,
		Code:             m.Code,
		Headers:          m.Headers,
		Body:             base64.StdEncoding.EncodeToString(m.Body),
		IsDynamic:        m.IsDynamic,
		CreatedAt:        m.CreatedAt,
	}
}

func (act *dbactions) HTTPRoutesCreate(ctx context.Context, p actions.HTTPRoutesCreateParams) (*actions.HTTPRoutesCreateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	if _, err := act.db.HTTPRoutesGetByPayloadMethodAndPath(ctx, payload.ID, strings.ToUpper(p.Method), p.Path); err != sql.ErrNoRows {
		return nil, errors.Conflictf("http route for payload %q with method %q and path %q already exist",
			p.PayloadName, strings.ToUpper(p.Method), p.Path)
	}

	body, err := base64.StdEncoding.DecodeString(p.Body)

	if err != nil {
		return nil, errors.Validationf("body: invalid base64 data")
	}

	rec := &models.HTTPRoute{
		PayloadID: payload.ID,
		Method:    strings.ToUpper(p.Method),
		Path:      p.Path,
		Code:      p.Code,
		Headers:   p.Headers,
		Body:      body,
		IsDynamic: p.IsDynamic,
	}

	if err := act.db.HTTPRoutesCreate(ctx, rec); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.HTTPRoutesCreateResult{HTTPRoute: HTTPRoute(*rec, payload.Subdomain)}, nil
}

func (act *dbactions) HTTPRoutesUpdate(ctx context.Context, p actions.HTTPRoutesUpdateParams) (*actions.HTTPRoutesUpdateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.Payload)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.Payload)
	}

	rec, err := act.db.HTTPRoutesGetByPayloadIDAndIndex(ctx, payload.ID, p.Index)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("http route for payload %q with index %d not found",
			p.Payload, p.Index)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if p.Method != nil {
		rec.Method = *p.Method
	}

	if p.Path != nil {
		rec.Path = *p.Path
	}

	if p.Code != nil {
		rec.Code = *p.Code
	}

	if p.Headers != nil {
		rec.Headers = p.Headers
	}

	if p.Body != nil {
		body, err := base64.StdEncoding.DecodeString(*p.Body)

		if err != nil {
			return nil, errors.Validationf("body: invalid base64 data")
		}

		rec.Body = body
	}

	if p.IsDynamic != nil {
		rec.IsDynamic = *p.IsDynamic
	}

	if err := act.db.HTTPRoutesUpdate(ctx, rec); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.HTTPRoutesUpdateResult{HTTPRoute: HTTPRoute(*rec, payload.Subdomain)}, nil
}

func (act *dbactions) HTTPRoutesDelete(ctx context.Context, p actions.HTTPRoutesDeleteParams) (*actions.HTTPRoutesDeleteResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	rec, err := act.db.HTTPRoutesGetByPayloadIDAndIndex(ctx, payload.ID, p.Index)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("http route for payload %q with index %d not found",
			p.PayloadName, p.Index)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.HTTPRoutesDelete(ctx, rec.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.HTTPRoutesDeleteResult{HTTPRoute: HTTPRoute(*rec, payload.Subdomain)}, nil
}

func (act *dbactions) HTTPRoutesClear(ctx context.Context, p actions.HTTPRoutesClearParams) (actions.HTTPRoutesClearResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	var recs []*models.HTTPRoute

	if p.Path != "" {
		recs, err = act.db.HTTPRoutesDeleteAllByPayloadIDAndPath(ctx, payload.ID, p.Path)
	} else {
		recs, err = act.db.HTTPRoutesDeleteAllByPayloadID(ctx, payload.ID)
	}
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.HTTPRoute, 0)

	for _, r := range recs {
		res = append(res, HTTPRoute(*r, payload.Subdomain))
	}

	return res, nil
}

func (act *dbactions) HTTPRoutesList(ctx context.Context, p actions.HTTPRoutesListParams) (actions.HTTPRoutesListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	recs, err := act.db.HTTPRoutesGetByPayloadID(ctx, payload.ID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.HTTPRoute, 0)

	for _, r := range recs {
		res = append(res, HTTPRoute(*r, payload.Subdomain))
	}

	return res, nil
}
