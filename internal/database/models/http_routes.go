package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var HTTPMethodAny = "ANY"

var HTTPMethods = []string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodTrace,
}

type HTTPRoute struct {
	Index     int64     `db:"index"`
	ID        int64     `db:"id"`
	PayloadID int64     `db:"payload_id"`
	Method    string    `db:"method"`
	Path      string    `db:"path"`
	Code      int       `db:"code"`
	Headers   Headers   `db:"headers"`
	Body      []byte    `db:"body"`
	IsDynamic bool      `db:"is_dynamic"`
	CreatedAt time.Time `db:"created_at"`
}

type Headers map[string][]string

func (m Headers) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Headers) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	*m = make(map[string][]string)

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	return nil
}
