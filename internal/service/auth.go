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

// LarkProvisionUser ensures a user exists for the given Lark ID, creating one on
// first contact. Lark is self-service: unlike the other messengers, a previously
// unknown Lark user is provisioned rather than rejected. Callers run it before
// AuthContextByLarkID, which stays lookup-only.
type LarkProvisionUser interface {
	LarkProvisionUser(context.Context, string) error
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
	LarkProvisionUser
}
