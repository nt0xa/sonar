package cmd

import (
	"context"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func GetUser(ctx context.Context) (*database.User, error) {
	u, ok := ctx.Value(userKey).(*database.User)
	if !ok {
		return nil, errors.Internalf("no %q key in context", userKey)
	}
	return u, nil
}

func SetUser(ctx context.Context, u *database.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}
