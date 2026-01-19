package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

type Event struct {
	Index      int64     `db:"index"`
	ID         int64     `db:"id"`
	UUID       uuid.UUID `db:"uuid"`
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

// Meta contains protocol-specific event metadata.
// Only one of the protocol-specific fields will be populated per event.
type Meta struct {
	// Protocol-specific metadata (only one is populated per event)
	DNS  *dnsx.Meta  `json:"dns,omitempty"`
	HTTP *httpx.Meta `json:"http,omitempty"`
	SMTP *smtpx.Meta `json:"smtp,omitempty"`
	FTP  *ftpx.Meta  `json:"ftp,omitempty"`

	// Common fields across protocols
	Secure bool `json:"secure,omitempty"`

	// GeoIP information (populated by event handler for all protocols)
	GeoIP *geoipx.Meta `json:"geoip,omitempty"`
}

func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Meta) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, m)
}
