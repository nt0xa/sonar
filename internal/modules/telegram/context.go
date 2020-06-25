package telegram

import (
	"context"

	"github.com/bi-zone/sonar/internal/utils/errors"
)

type contextKey string

const (
	chatIDKey contextKey = "telegram.chatID"
)

func GetChatID(ctx context.Context) (int64, errors.Error) {
	u, ok := ctx.Value(chatIDKey).(int64)
	if !ok {
		return 0, errors.Internalf("no %q key in context", chatIDKey)
	}
	return u, nil
}

func SetChatID(ctx context.Context, chatID int64) context.Context {
	return context.WithValue(ctx, chatIDKey, chatID)
}
