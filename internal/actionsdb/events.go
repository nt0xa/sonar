package actionsdb

import (
	"context"
	"database/sql"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func Event(m *models.Event) *actions.Event {
	if m == nil {
		return nil
	}

	return &actions.Event{
		Index:      m.Index,
		Protocol:   m.Protocol.String(),
		R:          "",
		W:          "",
		RW:         "",
		Meta:       m.Meta,
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

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.PayloadName)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.PayloadName)
	}

	recs, err := act.db.EventsListByPayloadID(payload.ID,
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

	res := make([]*actions.Event, 0)

	for _, r := range recs {
		res = append(res, Event(r))
	}

	return res, nil
}
