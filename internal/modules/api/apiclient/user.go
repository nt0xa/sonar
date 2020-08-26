package apiclient

import (
	"context"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (c *Client) UserCurrent(ctx context.Context) (actions.UserCurrentResult, errors.Error) {
	var res actions.UserCurrentResult

	resp, err := c.client.R().
		SetError(&apiError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/user")

	if err != nil {
		return nil, errors.Internal(err)
	}

	if resp.Error() != nil {
		return nil, resp.Error().(*apiError)
	}

	return res, nil
}
