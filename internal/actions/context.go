package actions

import (
	"context"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func GetUser(ctx context.Context) (*models.User, errors.Error) {
	u, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		return nil, errors.Internalf("no %q key in context", userKey)
	}
	return u, nil
}

func SetUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}
