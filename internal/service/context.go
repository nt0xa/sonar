package service

import "context"

// Caller is the authenticated identity behind a request. It is minted by the
// auth boundary (see dbsvc.AuthContextBy*) and carried through the request
// context. It holds everything downstream needs about the actor — including the
// name and social IDs — so consumers like auditsvc don't have to re-query the
// user from the database.
type Caller struct {
	UserID   int64
	UserName string
	IsAdmin  bool
	Source   AuditSource

	TelegramID *int64
	LarkID     *string
	SlackID    *string
}

type callerCtxKey struct{}

// WithCaller returns a context carrying c. It should only be called by the auth
// boundary; everything else reads identity via CallerFrom.
func WithCaller(ctx context.Context, c Caller) context.Context {
	return context.WithValue(ctx, callerCtxKey{}, c)
}

// CallerFrom returns the authenticated caller, if the context carries one.
func CallerFrom(ctx context.Context) (Caller, bool) {
	c, ok := ctx.Value(callerCtxKey{}).(Caller)
	return c, ok
}
