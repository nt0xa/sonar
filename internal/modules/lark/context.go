package lark

import (
	"context"

	"github.com/russtone/sonar/internal/utils/errors"
)

type contextKey string

const (
	messageIDKey contextKey = "lark.messageID"
)

func GetMessageID(ctx context.Context) (*string, errors.Error) {
	id, ok := ctx.Value(messageIDKey).(string)
	if !ok {
		return nil, errors.Internalf("no %q key in context", messageIDKey)
	}
	return &id, nil
}

func SetMessageID(ctx context.Context, msgID string) context.Context {
	return context.WithValue(ctx, messageIDKey, msgID)
}
