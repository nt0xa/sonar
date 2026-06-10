package service

import (
	"time"
)

//go:generate go-enum --ptr --names --values

// ENUM(dns, http, smtp, ftp)
type ProtoCategory string

// ENUM(dns, http, https, smtp, ftp)
type EventProtocol string

// ENUM(A, AAAA, MX, TXT, CNAME, NS, CAA)
type DNSRecordType string

// ENUM(all, round-robin, rebind)
type DNSRecordStrategy string

// ENUM(GET, HEAD, POST, PUT, PATCH, DELETE, CONNECT, OPTIONS, TRACE, ANY)
type HTTPMethod string

// ENUM(create, update, delete)
type AuditAction string

// ENUM(payload, user, dns_record, http_route)
type AuditResourceType string

// ENUM(api, telegram, lark, slack)
type AuditSource string

type Payload struct {
	Name            string          `json:"name"`
	Subdomain       string          `json:"subdomain"`
	NotifyProtocols []ProtoCategory `json:"notifyProtocols"`
	StoreEvents     bool            `json:"storeEvents"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type User struct {
	Name      string    `json:"name"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt"`

	APIToken   *string `json:"apiToken,omitempty"`
	TelegramID *int64  `json:"telegramId,omitempty"`
	LarkID     *string `json:"larkId,omitempty"`
	SlackID    *string `json:"slackId,omitempty"`
}

type DNSRecord struct {
	Index            int64             `json:"index"`
	PayloadSubdomain string            `json:"payloadSubdomain"`
	Name             string            `json:"name"`
	Type             DNSRecordType     `json:"type"`
	TTL              int               `json:"ttl"`
	Values           []string          `json:"values"`
	Strategy         DNSRecordStrategy `json:"strategy"`
	CreatedAt        time.Time         `json:"createdAt"`
}

type HTTPRoute struct {
	Index            int64               `json:"index"`
	PayloadSubdomain string              `json:"payloadSubdomain"`
	Method           HTTPMethod          `json:"method"`
	Path             string              `json:"path"`
	Code             int                 `json:"code"`
	Headers          map[string][]string `json:"headers"`
	Body             string              `json:"body"`
	IsDynamic        bool                `json:"isDynamic"`
	CreatedAt        time.Time           `json:"createdAt"`
}

type AuditRecord struct {
	ID           int64             `json:"id"`
	UUID         string            `json:"uuid"`
	CreatedAt    time.Time         `json:"createdAt"`
	Action       AuditAction       `json:"action"`
	ResourceType AuditResourceType `json:"resourceType"`
	Source       AuditSource       `json:"source"`
	ActorID      *int64            `json:"actorId,omitempty"`
	ActorName    string            `json:"actorName,omitempty"`
	ActorMeta    map[string]any    `json:"actorMetadata"`
	Resource     map[string]any    `json:"resource"`
}

type Event struct {
	Index      int64         `json:"index"`
	UUID       string        `json:"uuid"`
	Protocol   EventProtocol `json:"protocol"`
	R          string        `json:"r,omitempty"`
	W          string        `json:"w,omitempty"`
	RW         string        `json:"rw,omitempty"`
	Meta       EventMeta     `json:"meta"`
	RemoteAddr string        `json:"remoteAddress"`
	ReceivedAt time.Time     `json:"receivedAt"`
}

type EventMeta struct {
	DNS  *EventDNSMeta  `json:"dns,omitempty"`
	HTTP *EventHTTPMeta `json:"http,omitempty"`
	SMTP *EventSMTPMeta `json:"smtp,omitempty"`
	FTP  *EventFTPMeta  `json:"ftp,omitempty"`

	Secure bool `json:"secure"`

	GeoIP *EventGeoIPMeta `json:"geoip,omitempty"`
}

type EventHTTPRequest struct {
	Method  string              `json:"method"`
	Proto   string              `json:"proto"`
	URL     string              `json:"url"`
	Host    string              `json:"host"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type EventHTTPResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type EventHTTPMeta struct {
	Request  EventHTTPRequest  `json:"request"`
	Response EventHTTPResponse `json:"response"`
}

type EventDNSQuestion struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type EventDNSAnswer struct {
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  uint32 `json:"ttl"`
}

type EventDNSMeta struct {
	Question EventDNSQuestion `json:"question"`
	Answer   []EventDNSAnswer `json:"answer"`
}

type EventSMTPData struct {
	Helo     string   `json:"helo"`
	Ehlo     string   `json:"ehlo"`
	MailFrom string   `json:"mailFrom"`
	RcptTo   []string `json:"rcptTo"`
	Data     string   `json:"data"`
}

type EventSMTPAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type EventSMTPEmail struct {
	Subject string             `json:"subject"`
	From    []EventSMTPAddress `json:"from"`
	To      []EventSMTPAddress `json:"to"`
	Cc      []EventSMTPAddress `json:"cc"`
	Bcc     []EventSMTPAddress `json:"bcc"`
	Date    *time.Time         `json:"date,omitempty"`
	Text    string             `json:"text"`
	HTML    string             `json:"html"`
}

type EventSMTPMeta struct {
	Session EventSMTPData  `json:"session"`
	Email   EventSMTPEmail `json:"email"`
}

type EventFTPData struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Type string `json:"type"`
	Pasv string `json:"pasv"`
	Epsv string `json:"epsv"`
	Port string `json:"port"`
	Eprt string `json:"eprt"`
	Retr string `json:"retr"`
}

type EventFTPMeta struct {
	Session EventFTPData `json:"session"`
}

type EventGeoIPCountry struct {
	Name    string `json:"name"`
	ISOCode string `json:"isoCode"`
}

type EventGeoIPASN struct {
	Number uint   `json:"number"`
	Org    string `json:"org"`
}

type EventGeoIPMeta struct {
	City         string            `json:"city"`
	Country      EventGeoIPCountry `json:"country"`
	Subdivisions []string          `json:"subdivisions"`
	ASN          EventGeoIPASN     `json:"asn"`
}
