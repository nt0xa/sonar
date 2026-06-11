package service

import "context"

type userIDCtxKey struct{}

type sourceCtxKey struct{}

func SetUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, id)
}

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDCtxKey{}).(int64)
	return id, ok
}

func SetSource(ctx context.Context, src AuditSource) context.Context {
	return context.WithValue(ctx, sourceCtxKey{}, src)
}

func GetSource(ctx context.Context) (AuditSource, bool) {
	src, ok := ctx.Value(sourceCtxKey{}).(AuditSource)
	return src, ok
}
