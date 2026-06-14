package dbsvc

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

func payload(m database.Payload) *service.Payload {
	notifyProtocols := make([]service.ProtoCategory, len(m.NotifyProtocols))
	for i, p := range m.NotifyProtocols {
		notifyProtocols[i] = service.ProtoCategory(p)
	}

	return &service.Payload{
		Subdomain:       m.Subdomain,
		Name:            m.Name,
		NotifyProtocols: notifyProtocols,
		StoreEvents:     m.StoreEvents,
		CreatedAt:       m.CreatedAt,
	}
}
