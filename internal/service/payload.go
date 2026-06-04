package service

import (
	"time"
)

//go:generate go-enum --ptr --names --values

// ENUM(dns, http, smtp, ftp)
type ProtoCategory string

type Payload struct {
	Name            string
	Subdomain       string
	NotifyProtocols []ProtoCategory
	StoreEvents     bool
	CreatedAt       time.Time
}
