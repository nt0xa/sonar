package apiclient

import (
	"context"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Client) UsersCreate(ctx context.Context, p actions.UsersCreateParams) (actions.UsersCreateResult, errors.Error) {
	var res actions.UsersCreateResult

	resp, err := c.client.R().
		SetBody(p).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/users")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}

func (c *Client) UsersDelete(ctx context.Context, p actions.UsersDeleteParams) (actions.UsersDeleteResult, errors.Error) {
	var res actions.UsersDeleteResult

	resp, err := c.client.R().
		SetPathParams(toPath(p)).
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/users/{name}")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}
