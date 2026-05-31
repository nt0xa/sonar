package types

import (
	"time"
)

type Event struct {
	Index      int64
	UUID       string
	Protocol   string
	R          string
	W          string
	RW         string
	Meta       EventMeta
	RemoteAddr string
	ReceivedAt time.Time
}

type EventMeta struct {
	DNS  *EventDNSMeta
	HTTP *EventHTTPMeta
	SMTP *EventSMTPMeta
	FTP  *EventFTPMeta

	Secure bool

	GeoIP *EventGeoIPMeta
}

type EventHTTPRequest struct {
	Method  string
	Proto   string
	URL     string
	Host    string
	Headers map[string][]string
	Body    string
}

type EventHTTPResponse struct {
	Status  int
	Headers map[string][]string
	Body    string
}

type EventHTTPMeta struct {
	Request  EventHTTPRequest
	Response EventHTTPResponse
}

type EventDNSQuestion struct {
	Name string
	Type string
}

type EventDNSAnswer struct {
	Name string
	Type string
	TTL  uint32
}

type EventDNSMeta struct {
	Question EventDNSQuestion
	Answer   []EventDNSAnswer
}

type EventSMTPData struct {
	Helo     string
	Ehlo     string
	MailFrom string
	RcptTo   []string
	Data     string
}

type EventSMTPAddress struct {
	Name    string
	Address string
}

type EventSMTPEmail struct {
	Subject string
	From    []EventSMTPAddress
	To      []EventSMTPAddress
	Cc      []EventSMTPAddress
	Bcc     []EventSMTPAddress
	Date    *time.Time
	Text    string
	HTML    string
}

type EventSMTPMeta struct {
	Session EventSMTPData
	Email   EventSMTPEmail
}

type EventFTPData struct {
	User string
	Pass string
	Type string
	Pasv string
	Epsv string
	Port string
	Eprt string
	Retr string
}

type EventFTPMeta struct {
	Session EventFTPData
}

type EventGeoIPCountry struct {
	Name    string
	ISOCode string
}

type EventGeoIPASN struct {
	Number uint
	Org    string
}

type EventGeoIPMeta struct {
	City         string
	Country      EventGeoIPCountry
	Subdivisions []string
	ASN          EventGeoIPASN
}
