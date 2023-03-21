package actionsdb

import (
	"context"
	"database/sql"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/utils"
	"github.com/russtone/sonar/internal/utils/errors"
	"github.com/russtone/sonar/internal/utils/slice"
)

func Payload(m models.Payload) actions.Payload {
	return actions.Payload{
		Subdomain:       m.Subdomain,
		Name:            m.Name,
		NotifyProtocols: m.NotifyProtocols.Strings(),
		StoreEvents:     m.StoreEvents,
		CreatedAt:       m.CreatedAt,
	}
}

func (act *dbactions) PayloadsCreate(ctx context.Context, p actions.PayloadsCreateParams) (*actions.PayloadsCreateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	if _, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name); err != sql.ErrNoRows {
		return nil, errors.Conflictf("payload with name %q already exist", p.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, errors.Internal(err)
	}

	rec := &models.Payload{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            p.Name,
		NotifyProtocols: models.ProtoCategories(slice.StringsDedup(p.NotifyProtocols)...),
		StoreEvents:     p.StoreEvents,
	}

	err = act.db.PayloadsCreate(rec)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsCreateResult{Payload(*rec)}, nil
}

func (act *dbactions) PayloadsUpdate(ctx context.Context, p actions.PayloadsUpdateParams) (*actions.PayloadsUpdateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if p.NewName != "" {
		rec.Name = p.NewName
	}

	if p.NotifyProtocols != nil {
		rec.NotifyProtocols = models.ProtoCategories(slice.StringsDedup(p.NotifyProtocols)...)
	}

	if p.StoreEvents != nil {
		rec.StoreEvents = *p.StoreEvents
	}

	err = act.db.PayloadsUpdate(rec)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsUpdateResult{Payload(*rec)}, nil
}

func (act *dbactions) PayloadsDelete(ctx context.Context, p actions.PayloadsDeleteParams) (*actions.PayloadsDeleteResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("you don't have payload with name %q", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.PayloadsDelete(rec.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsDeleteResult{Payload(*rec)}, nil
}

func (act *dbactions) PayloadsList(ctx context.Context, p actions.PayloadsListParams) (actions.PayloadsListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	recs, err := act.db.PayloadsFindByUserAndName(u.ID, p.Name)
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.Payload, 0)

	for _, r := range recs {
		res = append(res, Payload(*r))
	}

	return res, nil
}
