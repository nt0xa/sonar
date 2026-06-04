package service

import (
	"time"
)

type User struct {
	Name      string
	IsAdmin   bool
	CreatedAt time.Time

	APIToken   *string
	TelegramID *int64
	LarkID     *string
	SlackID    *string
}
