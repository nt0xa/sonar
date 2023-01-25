package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/russtone/sonar/internal/actions"
)

var _ actions.ResultHandler = &jsonHandler{}

type jsonHandler struct {
	writer io.Writer
}

func (h *jsonHandler) json(data interface{}) {
	json.NewEncoder(h.writer).Encode(data)
}

//
// User
//

func (h *jsonHandler) UserCurrent(ctx context.Context, res actions.UserCurrentResult) {
	h.json(res)
}

//
// Payloads
//

func (h *jsonHandler) PayloadsCreate(ctx context.Context, res actions.PayloadsCreateResult) {
	h.json(res)
}

func (h *jsonHandler) PayloadsList(ctx context.Context, res actions.PayloadsListResult) {
	h.json(res)
}

func (h *jsonHandler) PayloadsUpdate(ctx context.Context, res actions.PayloadsUpdateResult) {
	h.json(res)
}

func (h *jsonHandler) PayloadsDelete(ctx context.Context, res actions.PayloadsDeleteResult) {
	h.json(res)
}

//
// DNS records
//

func (h *jsonHandler) DNSRecordsCreate(ctx context.Context, res actions.DNSRecordsCreateResult) {
	h.json(res)
}

func (h *jsonHandler) DNSRecordsList(ctx context.Context, res actions.DNSRecordsListResult) {
	h.json(res)
}

func (h *jsonHandler) DNSRecordsDelete(ctx context.Context, res actions.DNSRecordsDeleteResult) {
	h.json(res)
}

//
// HTTP routes
//

func (h *jsonHandler) HTTPRoutesCreate(ctx context.Context, res actions.HTTPRoutesCreateResult) {
	h.json(res)
}

func (h *jsonHandler) HTTPRoutesList(ctx context.Context, res actions.HTTPRoutesListResult) {
	h.json(res)
}

func (h *jsonHandler) HTTPRoutesDelete(ctx context.Context, res actions.HTTPRoutesDeleteResult) {
	h.json(res)
}

//
// Users
//

func (h *jsonHandler) UsersCreate(ctx context.Context, res actions.UsersCreateResult) {
	h.json(res)
}

func (h *jsonHandler) UsersDelete(ctx context.Context, res actions.UsersDeleteResult) {
	h.json(res)
}

//
// Events
//

func (h *jsonHandler) EventsList(ctx context.Context, res actions.EventsListResult) {
	h.json(res)
}

func (h *jsonHandler) EventsGet(ctx context.Context, res actions.EventsGetResult) {
	h.json(res)
}
