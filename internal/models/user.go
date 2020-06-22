package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/mitchellh/mapstructure"
)

type UserParamsKey string

const (
	UserTelegramID UserParamsKey = "telegram.id"
	UserAPIToken   UserParamsKey = "api.token"
)

type User struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	Params    UserParams `db:"params"`
	IsAdmin   bool       `db:"is_admin"`
	CreatedAt time.Time  `db:"created_at"`
}

type UserParams struct {
	TelegramID int64  `json:"telegram.id" mapstructure:"telegram.id"`
	APIToken   string `json:"api.token"   mapstructure:"api.token"`
}

func (p UserParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *UserParams) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	m := make(map[string]string)

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &p,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(c)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}
