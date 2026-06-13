package service

import "context"

type AuthContextByAPIToken interface {
	AuthContextByAPIToken(context.Context, string) (context.Context, error)
}

type AuthContextByTelegramID interface {
	AuthContextByTelegramID(context.Context, int64) (context.Context, error)
}

type AuthContextBySlackID interface {
	AuthContextBySlackID(context.Context, string) (context.Context, error)
}

type AuthContextByLarkID interface {
	AuthContextByLarkID(context.Context, string) (context.Context, error)
}

// ServerService is the full server-side service: the business operations of
// [Service] plus the ability to resolve a caller's identity into an
// authenticated context. It is implemented by DB-backed services (dbsvc and
// its decorators); the api-backed client (remotesvc) implements only [Service],
// since identity resolution happens on the server.
type ServerService interface {
	Service

	AuthContextByAPIToken
	AuthContextByTelegramID
	AuthContextBySlackID
	AuthContextByLarkID
}
