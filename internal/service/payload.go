package service

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

func payload(m database.Payload) *types.Payload {
	notifyProtocols := make([]types.ProtoCategory, len(m.NotifyProtocols))
	for i, p := range m.NotifyProtocols {
		notifyProtocols[i] = types.ProtoCategory(p)
	}

	return &types.Payload{
		Subdomain:       m.Subdomain,
		Name:            m.Name,
		NotifyProtocols: notifyProtocols,
		StoreEvents:     m.StoreEvents,
		CreatedAt:       m.CreatedAt,
	}
}
