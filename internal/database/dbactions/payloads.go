package dbactions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/errors"
	"github.com/bi-zone/sonar/internal/utils/slice"
)

func (act *dbactions) CreatePayload(ctx context.Context, p actions.CreatePayloadParams) (actions.CreatePayloadResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
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

	payload := &models.Payload{
		UserID:          u.ID,
		Subdomain:       subdomain,
		Name:            p.Name,
		NotifyProtocols: slice.StringsDedup(p.NotifyProtocols),
	}

	err = act.db.PayloadsCreate(payload)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return actions.CreatePayloadResult(payload), nil
}

func (act *dbactions) UpdatePayload(ctx context.Context, p actions.UpdatePayloadParams) (actions.UpdatePayloadResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("payload with name %q not found", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if p.NewName != "" {
		payload.Name = p.NewName
	}

	if p.NotifyProtocols != nil {
		payload.NotifyProtocols = slice.StringsDedup(p.NotifyProtocols)
	}

	err = act.db.PayloadsUpdate(payload)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return payload, nil
}

func (act *dbactions) DeletePayload(ctx context.Context, p actions.DeletePayloadParams) (actions.DeletePayloadResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payload, err := act.db.PayloadsGetByUserAndName(u.ID, p.Name)
	if err == sql.ErrNoRows {
		return nil, errors.NotFoundf("you don't have payload with name %q", p.Name)
	} else if err != nil {
		return nil, errors.Internal(err)
	}

	if err := act.db.PayloadsDelete(payload.ID); err != nil {
		return nil, errors.Internal(err)
	}

	return &actions.MessageResult{Message: fmt.Sprintf("payload %q deleted", payload.Name)}, nil
}

func (act *dbactions) ListPayloads(ctx context.Context, p actions.ListPayloadsParams) (actions.ListPayloadsResult, errors.Error) {
	u, e := actions.GetUser(ctx)
	if e != nil || u == nil {
		return nil, e
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Validation(err)
	}

	payloads, err := act.db.PayloadsFindByUserAndName(u.ID, p.Name)
	if err != nil {
		return nil, errors.Internal(err)
	}

	return actions.ListPayloadsResult(payloads), nil
}
