package actionsdb

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func Event(m models.Event) actions.Event {
	// Convert models.Meta to map[string]interface{} via JSON marshaling
	// This uses the custom MarshalJSON method which produces the correct format
	var meta map[string]interface{}
	if metaBytes, err := json.Marshal(m.Meta); err == nil {
		_ = json.Unmarshal(metaBytes, &meta)
	}

	return actions.Event{
		Index:      m.Index,
		Protocol:   m.Protocol.String(),
		R:          base64.StdEncoding.EncodeToString(m.R),
		W:          base64.StdEncoding.EncodeToString(m.W),
		RW:         base64.StdEncoding.EncodeToString(m.RW),
		Meta:       meta,
		RemoteAddr: m.RemoteAddr,
		ReceivedAt: m.ReceivedAt,
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
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	recs, err := act.db.EventsListByPayloadID(ctx, payload.ID,
		database.EventsPagination(database.Page{
			Count:  p.Count,
			After:  p.After,
			Before: p.Before,
		}),
		database.EventsReverse(p.Reverse),
	)
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.Event, 0)

	for _, r := range recs {
		res = append(res, Event(*r))
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
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	r, err := act.db.EventsGetByPayloadAndIndex(ctx, payload.ID, p.Index)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.EventsGetResult{Event: Event(*r)}, nil
}
