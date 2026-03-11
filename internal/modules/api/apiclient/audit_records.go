package apiclient

import (
	"context"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func (c *Client) AuditRecordsList(ctx context.Context, params actions.AuditRecordsListParams) (actions.AuditRecordsListResult, errors.Error) {
	var res actions.AuditRecordsListResult

	err := handle(c.client.R().
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/audit_records"))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) AuditRecordsGet(ctx context.Context, params actions.AuditRecordsGetParams) (*actions.AuditRecordsGetResult, errors.Error) {
	var res *actions.AuditRecordsGetResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/audit_records/{id}"))
	if err != nil {
		return nil, err
	}

	return res, nil
}
