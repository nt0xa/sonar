package types

import (
	"time"
)

//go:generate go-enum --ptr --names --values

// ENUM(GET, HEAD, POST, PUT, PATCH, DELETE, CONNECT, OPTIONS, TRACE, ANY)
type HTTPMethod string

type HTTPRoute struct {
	Index            int64
	PayloadSubdomain string
	Method           HTTPMethod
	Path             string
	Code             int
	Headers          map[string][]string
	Body             string
	IsDynamic        bool
	CreatedAt        time.Time
}
