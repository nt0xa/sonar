package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
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

// Meta contains protocol-specific metadata for events.
// Uses embedded struct pointers to maintain backward compatibility with JSON format.
// Only one embedded struct should be non-nil at a time based on the protocol.
type Meta struct {
	*DNSMeta
	*HTTPMeta
	*SMTPMeta
	*FTPMeta
	*GeoIPMeta
}

// DNSMeta contains DNS-specific metadata.
type DNSMeta struct {
	Question DNSQuestion `json:"question"`
	Answer   []DNSAnswer `json:"answer,omitempty"`
}

// DNSQuestion contains DNS query information.
type DNSQuestion struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// DNSAnswer contains DNS answer information.
type DNSAnswer struct {
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  uint32 `json:"ttl"`
}

// HTTPMeta contains HTTP-specific metadata.
type HTTPMeta struct {
	Request  HTTPRequest  `json:"request"`
	Response HTTPResponse `json:"response"`
	Secure   bool         `json:"secure"`
}

// HTTPRequest contains HTTP request information.
type HTTPRequest struct {
	Method  string              `json:"method"`
	Proto   string              `json:"proto"`
	URL     string              `json:"url"`
	Host    string              `json:"host"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// HTTPResponse contains HTTP response information.
type HTTPResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

// SMTPMeta contains SMTP-specific metadata.
type SMTPMeta struct {
	Session SMTPSession `json:"session"`
	Email   SMTPEmail   `json:"email"`
	Secure  bool        `json:"secure"`
}

// SMTPSession contains SMTP session information.
type SMTPSession struct {
	Helo     string   `json:"helo"`
	Ehlo     string   `json:"ehlo"`
	MailFrom string   `json:"mailFrom"`
	RcptTo   []string `json:"rcptTo"`
	Data     string   `json:"data"`
}

// SMTPEmail contains parsed email information.
type SMTPEmail struct {
	Subject string        `json:"subject"`
	From    []SMTPAddress `json:"from"`
	To      []SMTPAddress `json:"to"`
	Cc      []SMTPAddress `json:"cc"`
	Bcc     []SMTPAddress `json:"bcc"`
	Date    *time.Time    `json:"date,omitempty"`
	Text    string        `json:"text"`
}

// SMTPAddress contains email address information.
type SMTPAddress struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// FTPMeta contains FTP-specific metadata.
type FTPMeta struct {
	Session FTPSession `json:"session"`
	Secure  bool       `json:"secure"`
}

// FTPSession contains FTP session information.
type FTPSession struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Type string `json:"type"`
	Pasv string `json:"pasv"`
	Epsv string `json:"epsv"`
	Port string `json:"port"`
	Eprt string `json:"eprt"`
	Retr string `json:"retr"`
}

// GeoIPMeta contains geographic IP information.
type GeoIPMeta struct {
	City         string       `json:"city,omitempty"`
	Country      GeoIPCountry `json:"country,omitempty"`
	Subdivisions []string     `json:"subdivisions,omitempty"`
	ASN          GeoIPASN     `json:"asn,omitempty"`
}

// GeoIPCountry contains country information.
type GeoIPCountry struct {
	Name      string `json:"name,omitempty"`
	ISOCode   string `json:"isoCode,omitempty"`
	FlagEmoji string `json:"flagEmoji,omitempty"`
}

// GeoIPASN contains ASN information.
type GeoIPASN struct {
	Number uint   `json:"number,omitempty"`
	Org    string `json:"org,omitempty"`
}

func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Meta) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	if err := json.Unmarshal(b, m); err != nil {
		return err
	}

	return nil
}
