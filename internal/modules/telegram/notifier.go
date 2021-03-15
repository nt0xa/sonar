package telegram

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/bi-zone/sonar/internal/models"
)

const maxMessageSize = 4096

var (
	messageHeaderTemplate = tpl(`
<b>[{{ .Name }}]</b> {{ .Protocol }} from <code>{{ .RemoteAddr }}</code> at {{ .ReceivedAt }}
{{- if eq .Protocol "SMTP" }}
<b>Rcpt To:</b> {{ index .Meta "RcptTo" | join ", " }}
<b>Mail From:</b> {{ index .Meta "MailFrom" | join ", " }}
{{ end -}}
`)
	messageBodyTemplate = tpl(`
<pre>
{{ .Data }}
</pre>
`)
)

func (tg *Telegram) Notify(u *models.User, p *models.Payload, e *models.Event) error {

	var header, body bytes.Buffer

	headerData := struct {
		Name       string
		Protocol   string
		RemoteAddr string
		ReceivedAt string
		Meta       map[string]interface{}
	}{
		p.Name,
		e.Protocol.String(),
		e.RemoteAddr.String(),
		e.ReceivedAt.Format("15:04:05 01.01.2006"),
		e.Meta,
	}

	if err := messageHeaderTemplate.Execute(&header, headerData); err != nil {
		return fmt.Errorf("message header render error: %w", err)
	}

	bodyData := struct {
		Data string
	}{
		string(e.RawData),
	}

	if err := messageBodyTemplate.Execute(&body, bodyData); err != nil {
		return fmt.Errorf("message body render error: %w", err)
	}

	if len(header.String()+body.String()) < maxMessageSize && utf8.ValidString(body.String()) {
		tg.htmlMessage(u.Params.TelegramID, header.String()+body.String())
	} else {
		tg.docMessage(u.Params.TelegramID, "log.txt", header.String(), e.RawData)
	}

	// For SMTP send log.eml for better preview.
	if e.Protocol == "SMTP" {
		tg.docMessage(u.Params.TelegramID, "log.eml", header.String(), e.RawData)
	}

	return nil
}
