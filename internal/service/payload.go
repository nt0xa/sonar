package service

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

func payload(m database.Payload) *types.Payload {
	return &types.Payload{
		Subdomain:       m.Subdomain,
		Name:            m.Name,
		NotifyProtocols: m.NotifyProtocols,
		StoreEvents:     m.StoreEvents,
		CreatedAt:       m.CreatedAt,
	}
}
