package service

import "context"

type userIDCtxKey struct{}

func SetUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, id)
}

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDCtxKey{}).(int64)
	return id, ok
}

type userIsAdminCtxKey struct{}

func SetUserIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, userIsAdminCtxKey{}, isAdmin)
}

func GetUserIsAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(userIsAdminCtxKey{}).(bool)
	return ok && isAdmin
}

type sourceCtxKey struct{}

func SetSource(ctx context.Context, src AuditSource) context.Context {
	return context.WithValue(ctx, sourceCtxKey{}, src)
}

func GetSource(ctx context.Context) (AuditSource, bool) {
	src, ok := ctx.Value(sourceCtxKey{}).(AuditSource)
	return src, ok
}
