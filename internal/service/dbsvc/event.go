package dbsvc

import (
	"encoding/base64"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

func event(m database.Event, index int64) *service.Event {
	return &service.Event{
		Index:      index,
		UUID:       m.UUID.String(),
		Protocol:   service.EventProtocol(m.Protocol),
		R:          base64.StdEncoding.EncodeToString(m.R),
		W:          base64.StdEncoding.EncodeToString(m.W),
		RW:         base64.StdEncoding.EncodeToString(m.RW),
		Meta:       eventMeta(m.Meta),
		RemoteAddr: m.RemoteAddr,
		ReceivedAt: m.ReceivedAt,
	}
}

func eventMeta(m database.EventsMeta) service.EventMeta {
	out := service.EventMeta{
		Secure: m.Secure,
	}

	if m.DNS != nil {
		out.DNS = eventDNSMeta(*m.DNS)
	}
	if m.HTTP != nil {
		out.HTTP = eventHTTPMeta(*m.HTTP)
	}
	if m.SMTP != nil {
		out.SMTP = eventSMTPMeta(*m.SMTP)
	}
	if m.FTP != nil {
		out.FTP = eventFTPMeta(*m.FTP)
	}
	if m.GeoIP != nil {
		out.GeoIP = eventGeoIPMeta(*m.GeoIP)
	}

	return out
}

func eventDNSMeta(m dnsx.Meta) *service.EventDNSMeta {
	answers := make([]service.EventDNSAnswer, len(m.Answer))
	for i, a := range m.Answer {
		answers[i] = service.EventDNSAnswer{
			Name: a.Name,
			Type: a.Type,
			TTL:  a.TTL,
		}
	}

	return &service.EventDNSMeta{
		Question: service.EventDNSQuestion{
			Name: m.Question.Name,
			Type: m.Question.Type,
		},
		Answer: answers,
	}
}

func eventHTTPMeta(m httpx.Meta) *service.EventHTTPMeta {
	return &service.EventHTTPMeta{
		Request: service.EventHTTPRequest{
			Method:  m.Request.Method,
			Proto:   m.Request.Proto,
			URL:     m.Request.URL,
			Host:    m.Request.Host,
			Headers: m.Request.Headers,
			Body:    m.Request.Body,
		},
		Response: service.EventHTTPResponse{
			Status:  m.Response.Status,
			Headers: m.Response.Headers,
			Body:    m.Response.Body,
		},
	}
}

func eventSMTPMeta(m smtpx.Meta) *service.EventSMTPMeta {
	return &service.EventSMTPMeta{
		Session: service.EventSMTPData{
			Helo:     m.Session.Helo,
			Ehlo:     m.Session.Ehlo,
			MailFrom: m.Session.MailFrom,
			RcptTo:   m.Session.RcptTo,
			Data:     m.Session.Data,
		},
		Email: service.EventSMTPEmail{
			Subject: m.Email.Subject,
			From:    smtpAddresses(m.Email.From),
			To:      smtpAddresses(m.Email.To),
			Cc:      smtpAddresses(m.Email.Cc),
			Bcc:     smtpAddresses(m.Email.Bcc),
			Date:    m.Email.Date,
			Text:    m.Email.Text,
			HTML:    m.Email.HTML,
		},
	}
}

func smtpAddresses(in []smtpx.Address) []service.EventSMTPAddress {
	out := make([]service.EventSMTPAddress, len(in))
	for i, a := range in {
		out[i] = service.EventSMTPAddress{
			Name:    a.Name,
			Address: a.Address,
		}
	}
	return out
}

func eventFTPMeta(m ftpx.Meta) *service.EventFTPMeta {
	return &service.EventFTPMeta{
		Session: service.EventFTPData{
			User: m.Session.User,
			Pass: m.Session.Pass,
			Type: m.Session.Type,
			Pasv: m.Session.Pasv,
			Epsv: m.Session.Epsv,
			Port: m.Session.Port,
			Eprt: m.Session.Eprt,
			Retr: m.Session.Retr,
		},
	}
}

func eventGeoIPMeta(m geoipx.Meta) *service.EventGeoIPMeta {
	return &service.EventGeoIPMeta{
		City: m.City,
		Country: service.EventGeoIPCountry{
			Name:    m.Country.Name,
			ISOCode: m.Country.ISOCode,
		},
		Subdivisions: m.Subdivisions,
		ASN: service.EventGeoIPASN{
			Number: m.ASN.Number,
			Org:    m.ASN.Org,
		},
	}
}
