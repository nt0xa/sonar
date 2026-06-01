package service

import (
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
)

func user(m database.User) *types.User {
	return &types.User{
		Name:       m.Name,
		IsAdmin:    m.IsAdmin,
		CreatedAt:  m.CreatedAt,
		APIToken:   m.APIToken,
		TelegramID: m.TelegramID,
		LarkID:     m.LarkID,
		SlackID:    m.SlackID,
	}
}
