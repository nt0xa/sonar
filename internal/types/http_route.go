package types

import (
	"time"
)

type HTTPRoute struct {
	Index            int64
	PayloadSubdomain string
	Method           string
	Path             string
	Code             int
	Headers          map[string][]string
	Body             string
	IsDynamic        bool
	CreatedAt        time.Time
}
