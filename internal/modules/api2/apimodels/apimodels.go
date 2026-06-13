// Package apimodels holds the request payload types for the api2 HTTP API.
// They are shared between the server-side handlers and the api-backed
// [service.Service] client (remotesvc) so the wire contract stays in one place.
package apimodels

import "github.com/nt0xa/sonar/internal/service"

type PayloadsCreateRequest struct {
	Name            string                  `json:"name"`
	NotifyProtocols []service.ProtoCategory `json:"notifyProtocols"`
	StoreEvents     bool                    `json:"storeEvents"`
}

type PayloadsUpdateRequest struct {
	Name            string                  `json:"name"`
	NotifyProtocols []service.ProtoCategory `json:"notifyProtocols"`
	StoreEvents     *bool                   `json:"storeEvents"`
}

type DNSRecordsCreateRequest struct {
	PayloadName string                    `json:"payloadName"`
	Name        string                    `json:"name"`
	TTL         int                       `json:"ttl"`
	Type        service.DNSRecordType     `json:"type"`
	Values      []string                  `json:"values"`
	Strategy    service.DNSRecordStrategy `json:"strategy"`
}

type HTTPRoutesCreateRequest struct {
	PayloadName string              `json:"payloadName"`
	Method      service.HTTPMethod  `json:"method"`
	Path        string              `json:"path"`
	Code        int                 `json:"code"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	IsDynamic   bool                `json:"isDynamic"`
}

type HTTPRoutesUpdateRequest struct {
	Method    *service.HTTPMethod `json:"method"`
	Path      *string             `json:"path"`
	Code      *int                `json:"code"`
	Headers   map[string][]string `json:"headers"`
	Body      *string             `json:"body"`
	IsDynamic *bool               `json:"isDynamic"`
}

type UsersCreateRequest struct {
	Name       string  `json:"name"`
	APIToken   *string `json:"apiToken"`
	TelegramID *int64  `json:"telegramId"`
	LarkID     *string `json:"larkId"`
	SlackID    *string `json:"slackId"`
	IsAdmin    bool    `json:"isAdmin"`
}
