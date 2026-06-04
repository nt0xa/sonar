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
