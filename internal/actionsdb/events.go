package actionsdb

import (
	"context"
	"encoding/base64"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func Event(m database.Event, index int64) actions.Event {
	return actions.Event{
		Protocol:   m.Protocol,
		R:          base64.StdEncoding.EncodeToString(m.R),
		W:          base64.StdEncoding.EncodeToString(m.W),
		RW:         base64.StdEncoding.EncodeToString(m.RW),
		Meta:       m.Meta,
		RemoteAddr: m.RemoteAddr,
		ReceivedAt: m.ReceivedAt,
		UUID:       m.UUID.String(),
		Index:      index,
	}
}

func (act *dbactions) EventsList(ctx context.Context, p actions.EventsListParams) (actions.EventsListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == database.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	limit := p.Limit
	if limit == 0 {
		limit = 10
	}

	recs, err := act.db.EventsListByPayloadID(ctx, database.EventsListByPayloadIDParams{
		PayloadID: payload.ID,
		Limit:     int64(limit),
		Offset:    int64(p.Offset),
	})
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.Event, 0)

	for _, r := range recs {
		res = append(res, Event(r.Event, r.Index))
	}

	return res, nil
}

func (act *dbactions) EventsGet(ctx context.Context, p actions.EventsGetParams) (*actions.EventsGetResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.PayloadName)
	if err == database.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	r, err := act.db.EventsGetByPayloadAndIndex(ctx, payload.ID, p.Index)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.EventsGetResult{Event: Event(r.Event, r.Index)}, nil
}
