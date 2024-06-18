package apiclient

import (
	"context"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/utils/errors"
)

func (c *Client) DNSRecordsClear(ctx context.Context, params actions.DNSRecordsClearParams) (actions.DNSRecordsClearResult, errors.Error) {
	var res actions.DNSRecordsClearResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/dns-records/{payload}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) DNSRecordsCreate(ctx context.Context, params actions.DNSRecordsCreateParams) (*actions.DNSRecordsCreateResult, errors.Error) {
	var res *actions.DNSRecordsCreateResult

	err := handle(c.client.R().
		SetBody(params).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/dns-records"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) DNSRecordsDelete(ctx context.Context, params actions.DNSRecordsDeleteParams) (*actions.DNSRecordsDeleteResult, errors.Error) {
	var res *actions.DNSRecordsDeleteResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/dns-records/{payload}/{index}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) DNSRecordsList(ctx context.Context, params actions.DNSRecordsListParams) (actions.DNSRecordsListResult, errors.Error) {
	var res actions.DNSRecordsListResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/dns-records/{payload}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) EventsGet(ctx context.Context, params actions.EventsGetParams) (*actions.EventsGetResult, errors.Error) {
	var res *actions.EventsGetResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/events/{payload}/{index}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) EventsList(ctx context.Context, params actions.EventsListParams) (actions.EventsListResult, errors.Error) {
	var res actions.EventsListResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/events/{payload}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) HTTPRoutesClear(ctx context.Context, params actions.HTTPRoutesClearParams) (actions.HTTPRoutesClearResult, errors.Error) {
	var res actions.HTTPRoutesClearResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/http-routes/{payload}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) HTTPRoutesCreate(ctx context.Context, params actions.HTTPRoutesCreateParams) (*actions.HTTPRoutesCreateResult, errors.Error) {
	var res *actions.HTTPRoutesCreateResult

	err := handle(c.client.R().
		SetBody(params).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/http-routes"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) HTTPRoutesDelete(ctx context.Context, params actions.HTTPRoutesDeleteParams) (*actions.HTTPRoutesDeleteResult, errors.Error) {
	var res *actions.HTTPRoutesDeleteResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/http-routes/{payload}/{index}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) HTTPRoutesList(ctx context.Context, params actions.HTTPRoutesListParams) (actions.HTTPRoutesListResult, errors.Error) {
	var res actions.HTTPRoutesListResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/http-routes/{payload}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PayloadsClear(ctx context.Context, params actions.PayloadsClearParams) (actions.PayloadsClearResult, errors.Error) {
	var res actions.PayloadsClearResult

	err := handle(c.client.R().
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/payloads"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PayloadsCreate(ctx context.Context, params actions.PayloadsCreateParams) (*actions.PayloadsCreateResult, errors.Error) {
	var res *actions.PayloadsCreateResult

	err := handle(c.client.R().
		SetBody(params).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/payloads"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PayloadsDelete(ctx context.Context, params actions.PayloadsDeleteParams) (*actions.PayloadsDeleteResult, errors.Error) {
	var res *actions.PayloadsDeleteResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/payloads/{name}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PayloadsList(ctx context.Context, params actions.PayloadsListParams) (actions.PayloadsListResult, errors.Error) {
	var res actions.PayloadsListResult

	err := handle(c.client.R().
		SetQueryParamsFromValues(toQuery(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/payloads"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PayloadsUpdate(ctx context.Context, params actions.PayloadsUpdateParams) (*actions.PayloadsUpdateResult, errors.Error) {
	var res *actions.PayloadsUpdateResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetBody(params).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Put("/payloads/{name}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) ProfileGet(ctx context.Context) (*actions.ProfileGetResult, errors.Error) {
	var res *actions.ProfileGetResult

	err := handle(c.client.R().SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Get("/profile"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) UsersCreate(ctx context.Context, params actions.UsersCreateParams) (*actions.UsersCreateResult, errors.Error) {
	var res *actions.UsersCreateResult

	err := handle(c.client.R().
		SetBody(params).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Post("/users"))

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) UsersDelete(ctx context.Context, params actions.UsersDeleteParams) (*actions.UsersDeleteResult, errors.Error) {
	var res *actions.UsersDeleteResult

	err := handle(c.client.R().
		SetPathParams(toPath(params)).
		SetError(&APIError{}).
		SetResult(&res).
		SetContext(ctx).
		Delete("/users/{name}"))

	if err != nil {
		return nil, err
	}

	return res, nil
}
