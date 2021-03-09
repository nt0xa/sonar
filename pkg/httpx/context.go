package httpx

import (
	"context"
	"net"
	"net/http"
)

type contextKey struct {
	key string
}

var connKey = contextKey{"http-conn"}

func saveConn(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, connKey, c)
}

func getConn(r *http.Request) net.Conn {
	return r.Context().Value(connKey).(net.Conn)
}
