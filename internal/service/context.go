package service

import (
	"context"

	"github.com/nt0xa/sonar/internal/database"
)

type userCtxKey struct{}

func setUser(ctx context.Context, u database.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

func getUser(ctx context.Context) *database.User {
	u, ok := ctx.Value(userCtxKey{}).(database.User)
	if !ok {
		return nil
	}
	return &u
}
