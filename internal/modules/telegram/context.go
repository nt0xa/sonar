package telegram

import (
	"context"
	"errors"
)

type contextKey struct{}

type messageInfo struct {
	chatID int64
	msgID  int
}

func getMsgInfo(ctx context.Context) (*messageInfo, error) {
	mi, ok := ctx.Value(contextKey{}).(*messageInfo)
	if !ok {
		return nil, errors.New("no key in context")
	}
	return mi, nil
}

func setMsgInfo(ctx context.Context, chatID int64, msgID int) context.Context {
	return context.WithValue(ctx, contextKey{}, &messageInfo{chatID: chatID, msgID: msgID})
}
