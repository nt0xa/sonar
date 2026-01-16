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

// Meta represents event metadata with protocol-specific fields.
type Meta struct {
	// DNS specific metadata
	DNS *DNSMeta

	// HTTP/HTTPS specific metadata
	HTTP *HTTPMeta

	// SMTP specific metadata
	SMTP *SMTPMeta

	// FTP specific metadata
	FTP *FTPMeta

	// GeoIP metadata (common for all protocols)
	GeoIP *GeoIPMeta

	// For backward compatibility: store unmapped fields
	Extra map[string]interface{}
}

// DNSQuestion represents a DNS question
type DNSQuestion struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// DNSAnswer represents a DNS answer
type DNSAnswer struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	TTL  uint32 `json:"ttl,omitempty"`
}

// DNSMeta represents DNS-specific metadata
type DNSMeta struct {
	Question *DNSQuestion `json:"question,omitempty"`
	Answer   []DNSAnswer  `json:"answer,omitempty"`
}

// HTTPRequest represents HTTP request metadata
type HTTPRequest struct {
	Method  string      `json:"method,omitempty"`
	Proto   string      `json:"proto,omitempty"`
	URL     string      `json:"url,omitempty"`
	Host    string      `json:"host,omitempty"`
	Headers interface{} `json:"headers,omitempty"` // http.Header
	Body    string      `json:"body,omitempty"`
}

// HTTPResponse represents HTTP response metadata
type HTTPResponse struct {
	Status  int         `json:"status,omitempty"`
	Headers interface{} `json:"headers,omitempty"` // http.Header
	Body    string      `json:"body,omitempty"`
}

// HTTPMeta represents HTTP/HTTPS-specific metadata
type HTTPMeta struct {
	Request  HTTPRequest  `json:"request,omitempty"`
	Response HTTPResponse `json:"response,omitempty"`
	Secure   bool         `json:"secure,omitempty"`
}

// SMTPSession represents SMTP session data
type SMTPSession struct {
	Helo     string   `json:"helo,omitempty"`
	Ehlo     string   `json:"ehlo,omitempty"`
	MailFrom string   `json:"mailFrom,omitempty"`
	RcptTo   []string `json:"rcptTo,omitempty"`
	Data     string   `json:"data,omitempty"`
}

// SMTPAddress represents an email address
type SMTPAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// SMTPEmail represents parsed email data
type SMTPEmail struct {
	Subject string        `json:"subject,omitempty"`
	From    []SMTPAddress `json:"from,omitempty"`
	To      []SMTPAddress `json:"to,omitempty"`
	Cc      []SMTPAddress `json:"cc,omitempty"`
	Bcc     []SMTPAddress `json:"bcc,omitempty"`
	Date    *time.Time    `json:"date,omitempty"`
	Text    string        `json:"text,omitempty"`
}

// SMTPMeta represents SMTP-specific metadata
type SMTPMeta struct {
	Session SMTPSession `json:"session,omitempty"`
	Email   SMTPEmail   `json:"email,omitempty"`
	Secure  bool        `json:"secure,omitempty"`
}

// FTPSession represents FTP session data
type FTPSession struct {
	User string `json:"user,omitempty"`
	Pass string `json:"pass,omitempty"`
	Type string `json:"type,omitempty"`
	Pasv string `json:"pasv,omitempty"`
	Epsv string `json:"epsv,omitempty"`
	Port string `json:"port,omitempty"`
	Eprt string `json:"eprt,omitempty"`
	Retr string `json:"retr,omitempty"`
}

// FTPMeta represents FTP-specific metadata
type FTPMeta struct {
	Session FTPSession `json:"session,omitempty"`
	Secure  bool       `json:"secure,omitempty"`
}

// GeoIPCountry represents country information
type GeoIPCountry struct {
	Name      string `json:"name,omitempty"`
	ISOCode   string `json:"isoCode,omitempty"`
	FlagEmoji string `json:"flagEmoji,omitempty"`
}

// GeoIPASN represents ASN information
type GeoIPASN struct {
	Number uint   `json:"number,omitempty"`
	Org    string `json:"org,omitempty"`
}

// GeoIPMeta represents GeoIP metadata
type GeoIPMeta struct {
	City         string       `json:"city,omitempty"`
	Country      GeoIPCountry `json:"country,omitempty"`
	Subdivisions []string     `json:"subdivisions,omitempty"`
	ASN          GeoIPASN     `json:"asn,omitempty"`
}

// MarshalJSON implements custom JSON marshaling to maintain backward compatibility
func (m Meta) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})

	// Add DNS metadata
	if m.DNS != nil {
		if m.DNS.Question != nil {
			result["question"] = m.DNS.Question
		}
		if len(m.DNS.Answer) > 0 {
			result["answer"] = m.DNS.Answer
		}
	}

	// Add HTTP metadata
	if m.HTTP != nil {
		result["request"] = m.HTTP.Request
		result["response"] = m.HTTP.Response
		result["secure"] = m.HTTP.Secure
	}

	// Add SMTP metadata
	if m.SMTP != nil {
		result["session"] = m.SMTP.Session
		result["email"] = m.SMTP.Email
		result["secure"] = m.SMTP.Secure
	}

	// Add FTP metadata
	if m.FTP != nil {
		result["session"] = m.FTP.Session
		result["secure"] = m.FTP.Secure
	}

	// Add GeoIP metadata
	if m.GeoIP != nil {
		result["geoip"] = m.GeoIP
	}

	// Add extra fields
	for k, v := range m.Extra {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return json.Marshal(result)
}

// UnmarshalJSON implements custom JSON unmarshaling for backward compatibility
func (m *Meta) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.Extra = make(map[string]interface{})

	// Parse DNS metadata
	if question, ok := raw["question"].(map[string]interface{}); ok {
		m.DNS = &DNSMeta{
			Question: &DNSQuestion{
				Name: getString(question, "name"),
				Type: getString(question, "type"),
			},
		}
		if answer, ok := raw["answer"].([]interface{}); ok {
			var answers []DNSAnswer
			answerBytes, _ := json.Marshal(answer)
			_ = json.Unmarshal(answerBytes, &answers)
			m.DNS.Answer = answers
		}
		delete(raw, "question")
		delete(raw, "answer")
	}

	// Parse HTTP metadata
	if request, ok := raw["request"].(map[string]interface{}); ok {
		response, _ := raw["response"].(map[string]interface{})
		m.HTTP = &HTTPMeta{
			Request: HTTPRequest{
				Method:  getString(request, "method"),
				Proto:   getString(request, "proto"),
				URL:     getString(request, "url"),
				Host:    getString(request, "host"),
				Headers: request["headers"],
				Body:    getString(request, "body"),
			},
			Response: HTTPResponse{
				Status:  getInt(response, "status"),
				Headers: response["headers"],
				Body:    getString(response, "body"),
			},
			Secure: getBool(raw, "secure"),
		}
		delete(raw, "request")
		delete(raw, "response")
		delete(raw, "secure")
	}

	// Parse SMTP metadata
	if session, ok := raw["session"].(map[string]interface{}); ok {
		if email, ok := raw["email"].(map[string]interface{}); ok {
			// This is SMTP (has both session and email)
			m.SMTP = &SMTPMeta{
				Session: SMTPSession{
					Helo:     getString(session, "helo"),
					Ehlo:     getString(session, "ehlo"),
					MailFrom: getString(session, "mailFrom"),
					RcptTo:   getStringSlice(session, "rcptTo"),
					Data:     getString(session, "data"),
				},
				Secure: getBool(raw, "secure"),
			}

			// Parse email addresses
			fromAddrs := parseAddresses(email, "from")
			toAddrs := parseAddresses(email, "to")
			ccAddrs := parseAddresses(email, "cc")
			bccAddrs := parseAddresses(email, "bcc")

			m.SMTP.Email = SMTPEmail{
				Subject: getString(email, "subject"),
				From:    fromAddrs,
				To:      toAddrs,
				Cc:      ccAddrs,
				Bcc:     bccAddrs,
				Text:    getString(email, "text"),
			}
			// Parse date if present
			if dateStr, ok := email["date"].(string); ok {
				if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
					m.SMTP.Email.Date = &t
				}
			}

			delete(raw, "session")
			delete(raw, "email")
			delete(raw, "secure")
		} else {
			// This is FTP (has session but no email)
			m.FTP = &FTPMeta{
				Session: FTPSession{
					User: getString(session, "user"),
					Pass: getString(session, "pass"),
					Type: getString(session, "type"),
					Pasv: getString(session, "pasv"),
					Epsv: getString(session, "epsv"),
					Port: getString(session, "port"),
					Eprt: getString(session, "eprt"),
					Retr: getString(session, "retr"),
				},
				Secure: getBool(raw, "secure"),
			}
			delete(raw, "session")
			delete(raw, "secure")
		}
	}

	// Parse GeoIP metadata
	if geoip, ok := raw["geoip"].(map[string]interface{}); ok {
		m.GeoIP = &GeoIPMeta{
			City:         getString(geoip, "city"),
			Subdivisions: getStringSlice(geoip, "subdivisions"),
		}
		if country, ok := geoip["country"].(map[string]interface{}); ok {
			m.GeoIP.Country = GeoIPCountry{
				Name:      getString(country, "name"),
				ISOCode:   getString(country, "isoCode"),
				FlagEmoji: getString(country, "flagEmoji"),
			}
		}
		if asn, ok := geoip["asn"].(map[string]interface{}); ok {
			m.GeoIP.ASN = GeoIPASN{
				Number: getUint(asn, "number"),
				Org:    getString(asn, "org"),
			}
		}
		delete(raw, "geoip")
	}

	// Store remaining fields in Extra
	for k, v := range raw {
		m.Extra[k] = v
	}

	return nil
}

// Helper functions for unmarshaling
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func getUint(m map[string]interface{}, key string) uint {
	if v, ok := m[key].(float64); ok {
		return uint(v)
	}
	if v, ok := m[key].(int); ok {
		return uint(v)
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

func parseAddresses(m map[string]interface{}, key string) []SMTPAddress {
	if v, ok := m[key].([]interface{}); ok {
		result := make([]SMTPAddress, 0, len(v))
		for _, item := range v {
			if addr, ok := item.(map[string]interface{}); ok {
				result = append(result, SMTPAddress{
					Name:  getString(addr, "name"),
					Email: getString(addr, "email"),
				})
			}
		}
		return result
	}
	return nil
}

func (m Meta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Meta) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, m)
}
