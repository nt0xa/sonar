package dbsvc

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

func user(m database.User) *service.User {
	return &service.User{
		Name:       m.Name,
		IsAdmin:    m.IsAdmin,
		CreatedAt:  m.CreatedAt,
		APIToken:   m.APIToken,
		TelegramID: m.TelegramID,
		LarkID:     m.LarkID,
		SlackID:    m.SlackID,
	}
}
