package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Event struct {
	ID         int64     `db:"id"`
	PayloadID  int64     `db:"payload_id"`
	Protocol   Proto     `db:"protocol"`
	R          []byte    `db:"r"`
	W          []byte    `db:"w"`
	RW         []byte    `db:"rw"`
	Meta       Meta      `db:"meta"`
	RemoteAddr string    `db:"remote_addr"`
	ReceivedAt time.Time `db:"received_at"`
	CreatedAt  time.Time `db:"created_at"`
}

type Meta map[string]interface{}

func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Meta) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	*m = make(map[string]interface{})

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	return nil
}
