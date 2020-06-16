package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	Params    UserParams `db:"params"`
	CreatedAt time.Time  `db:"created_at"`
}

// It is required to add omitempty because of UsersGetByParams func
type UserParams struct {
	Admin      bool   `json:"admin,omitempty"       mapstructure:"admin"`
	TelegramID int64  `json:"telegram.id,omitempty" mapstructure:"telegram.id"`
	APIToken   string `json:"api.token,omitempty"   mapstructure:"api.token"`
}

func (p UserParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *UserParams) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}
