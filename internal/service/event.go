package service

import (
	"encoding/base64"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/types"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

func event(m database.Event, index int64) *types.Event {
	return &types.Event{
		Index:      index,
		UUID:       m.UUID.String(),
		Protocol:   types.EventProtocol(m.Protocol),
		R:          base64.StdEncoding.EncodeToString(m.R),
		W:          base64.StdEncoding.EncodeToString(m.W),
		RW:         base64.StdEncoding.EncodeToString(m.RW),
		Meta:       eventMeta(m.Meta),
		RemoteAddr: m.RemoteAddr,
		ReceivedAt: m.ReceivedAt,
	}
}

func eventMeta(m database.EventsMeta) types.EventMeta {
	out := types.EventMeta{
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

func eventDNSMeta(m dnsx.Meta) *types.EventDNSMeta {
	answers := make([]types.EventDNSAnswer, len(m.Answer))
	for i, a := range m.Answer {
		answers[i] = types.EventDNSAnswer{
			Name: a.Name,
			Type: a.Type,
			TTL:  a.TTL,
		}
	}

	return &types.EventDNSMeta{
		Question: types.EventDNSQuestion{
			Name: m.Question.Name,
			Type: m.Question.Type,
		},
		Answer: answers,
	}
}

func eventHTTPMeta(m httpx.Meta) *types.EventHTTPMeta {
	return &types.EventHTTPMeta{
		Request: types.EventHTTPRequest{
			Method:  m.Request.Method,
			Proto:   m.Request.Proto,
			URL:     m.Request.URL,
			Host:    m.Request.Host,
			Headers: m.Request.Headers,
			Body:    m.Request.Body,
		},
		Response: types.EventHTTPResponse{
			Status:  m.Response.Status,
			Headers: m.Response.Headers,
			Body:    m.Response.Body,
		},
	}
}

func eventSMTPMeta(m smtpx.Meta) *types.EventSMTPMeta {
	return &types.EventSMTPMeta{
		Session: types.EventSMTPData{
			Helo:     m.Session.Helo,
			Ehlo:     m.Session.Ehlo,
			MailFrom: m.Session.MailFrom,
			RcptTo:   m.Session.RcptTo,
			Data:     m.Session.Data,
		},
		Email: types.EventSMTPEmail{
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

func smtpAddresses(in []smtpx.Address) []types.EventSMTPAddress {
	out := make([]types.EventSMTPAddress, len(in))
	for i, a := range in {
		out[i] = types.EventSMTPAddress{
			Name:    a.Name,
			Address: a.Address,
		}
	}
	return out
}

func eventFTPMeta(m ftpx.Meta) *types.EventFTPMeta {
	return &types.EventFTPMeta{
		Session: types.EventFTPData{
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

func eventGeoIPMeta(m geoipx.Meta) *types.EventGeoIPMeta {
	return &types.EventGeoIPMeta{
		City: m.City,
		Country: types.EventGeoIPCountry{
			Name:    m.Country.Name,
			ISOCode: m.Country.ISOCode,
		},
		Subdivisions: m.Subdivisions,
		ASN: types.EventGeoIPASN{
			Number: m.ASN.Number,
			Org:    m.ASN.Org,
		},
	}
}
