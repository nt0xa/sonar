package actionsdb

import (
	"context"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/internal/utils/slice"
)

func Payload(m database.Payload) actions.Payload {
	return actions.Payload{
		Subdomain:       m.Subdomain,
		Name:            m.Name,
		NotifyProtocols: m.NotifyProtocols,
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

	if _, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.Name); err != database.ErrNoRows {
		return nil, errors.Conflictf("payload with name %q already exist", p.Name)
	}

	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return nil, errors.Internal(err)
	}

	rec, err := act.db.PayloadsCreate(ctx, database.PayloadsCreateParams{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            p.Name,
		NotifyProtocols: slice.StringsDedup(p.NotifyProtocols),
		StoreEvents:     p.StoreEvents,
	})
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsCreateResult{Payload: Payload(*rec)}, nil
}

func (act *dbactions) PayloadsUpdate(ctx context.Context, p actions.PayloadsUpdateParams) (*actions.PayloadsUpdateResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.Name)
	if err == database.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	name := rec.Name
	if p.NewName != "" {
		name = p.NewName
	}

	notifyProtocols := rec.NotifyProtocols
	if p.NotifyProtocols != nil {
		notifyProtocols = slice.StringsDedup(p.NotifyProtocols)
	}

	storeEvents := rec.StoreEvents
	if p.StoreEvents != nil {
		storeEvents = *p.StoreEvents
	}

	updated, err := act.db.PayloadsUpdate(ctx, database.PayloadsUpdateParams{
		ID:              rec.ID,
		UserID:          rec.UserID,
		Subdomain:       rec.Subdomain,
		Name:            name,
		NotifyProtocols: notifyProtocols,
		StoreEvents:     storeEvents,
	})
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsUpdateResult{Payload: Payload(*updated)}, nil
}

func (act *dbactions) PayloadsDelete(ctx context.Context, p actions.PayloadsDeleteParams) (*actions.PayloadsDeleteResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	rec, err := act.db.PayloadsGetByUserAndName(ctx, u.ID, p.Name)
	if err == database.ErrNoRows {
		return nil, errors.NotFoundf("you don't have payload with name %q", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	deleted, err := act.db.PayloadsDelete(ctx, rec.ID)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.PayloadsDeleteResult{Payload: Payload(*deleted)}, nil
}

func (act *dbactions) PayloadsClear(ctx context.Context, p actions.PayloadsClearParams) (actions.PayloadsClearResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	recs, err := act.db.PayloadsDeleteByNamePart(ctx, u.ID, p.Name)
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.Payload, 0)

	for _, r := range recs {
		res = append(res, Payload(*r))
	}

	return res, nil
}

func (act *dbactions) PayloadsList(ctx context.Context, p actions.PayloadsListParams) (actions.PayloadsListResult, errors.Error) {
	u, err := GetUser(ctx)
	if err != nil {
		return nil, errors.Internal(err)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	perPage := p.PerPage
	if perPage == 0 {
		perPage = 10
	}
	page := p.Page
	if page == 0 {
		page = 1
	}

	recs, err := act.db.PayloadsFindByUserAndName(ctx, database.PayloadsFindByUserAndNameParams{
		UserID: u.ID,
		Name:   p.Name,
		Limit:  int64(perPage),
		Offset: int64((page - 1) * perPage),
	})
	if err != nil {
		return nil, errors.Internal(err)
	}

	res := make([]actions.Payload, 0)

	for _, r := range recs {
		res = append(res, Payload(*r))
	}

	return res, nil
}
