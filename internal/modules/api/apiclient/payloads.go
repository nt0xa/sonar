package apiclient

import (
	"context"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Client) PayloadsCreate(ctx context.Context, p actions.PayloadsCreateParams) (actions.PayloadsCreateResult, errors.Error) {
	var res actions.PayloadsCreateResult

	resp, err := c.client.R().
		SetBody(p).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/payloads")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) PayloadsList(ctx context.Context, p actions.PayloadsListParams) (actions.PayloadsListResult, errors.Error) {
	var res actions.PayloadsListResult

	resp, err := c.client.R().
		SetQueryParamsFromValues(toQuery(p)).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/payloads")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) PayloadsUpdate(ctx context.Context, p actions.PayloadsUpdateParams) (actions.PayloadsUpdateResult, errors.Error) {
	var res actions.PayloadsUpdateResult

	resp, err := c.client.R().
		SetPathParams(toPath(p)).
		SetBody(p).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Put("/payloads/{name}")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) PayloadsDelete(ctx context.Context, p actions.PayloadsDeleteParams) (actions.PayloadsDeleteResult, errors.Error) {
	var res actions.PayloadsDeleteResult

	resp, err := c.client.R().
		SetPathParams(toPath(p)).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/payloads/{name}")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}
