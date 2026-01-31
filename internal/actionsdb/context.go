package actionsdb

import (
	"context"
	"fmt"

	"github.com/nt0xa/sonar/internal/database"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func SetUser(ctx context.Context, u *database.UsersFull) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func GetUser(ctx context.Context) (*database.UsersFull, error) {
	u, ok := ctx.Value(userKey).(*database.UsersFull)
	if !ok {
		return nil, fmt.Errorf("no %q key in context", userKey)
	}

	return u, nil
}
