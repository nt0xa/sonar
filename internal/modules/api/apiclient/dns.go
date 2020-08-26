package apiclient

import (
	"context"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Client) DNSRecordsCreate(ctx context.Context, p actions.DNSRecordsCreateParams) (actions.DNSRecordsCreateResult, errors.Error) {
	var res actions.DNSRecordsCreateResult

	resp, err := c.client.R().
		SetBody(p).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/dns")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) DNSRecordsList(ctx context.Context, p actions.DNSRecordsListParams) (actions.DNSRecordsListResult, errors.Error) {
	var res actions.DNSRecordsListResult

	resp, err := c.client.R().
		SetPathParams(toPath(p)).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/dns/{payloadName}")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) DNSRecordsDelete(ctx context.Context, p actions.DNSRecordsDeleteParams) (actions.DNSRecordsDeleteResult, errors.Error) {
	var res actions.DNSRecordsDeleteResult

	resp, err := c.client.R().
		SetPathParams(toPath(p)).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/dns/{payloadName}/{name}/{type}")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}
