package actionsdb

import (
	"context"
	"fmt"

	"github.com/nt0xa/sonar/internal/database/models"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func SetUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func GetUser(ctx context.Context) (*models.User, error) {
	u, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		return nil, fmt.Errorf("no %q key in context", userKey)
	}

	return u, nil
}
