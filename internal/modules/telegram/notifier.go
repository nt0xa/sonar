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

{{- if eq .Protocol "smtp" }}
<b>Rcpt To:</b> {{ index (index .Meta "session") "rcptTo" | join ", " }}
<b>Mail From:</b> {{ index (index .Meta "session") "mailFrom" | join ", " }}
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
		e.RemoteAddr,
		e.ReceivedAt.Format("15:04:05 01.01.2006"),
		e.Meta,
	}

	if err := messageHeaderTemplate.Execute(&header, headerData); err != nil {
		fmt.Println(err)
		return fmt.Errorf("message header render error: %w", err)
	}

	bodyData := struct {
		Data string
	}{
		string(e.RW),
	}

	if err := messageBodyTemplate.Execute(&body, bodyData); err != nil {
		return fmt.Errorf("message body render error: %w", err)
	}

	if len(header.String()+body.String()) < maxMessageSize && utf8.ValidString(body.String()) {
		tg.htmlMessage(u.Params.TelegramID, header.String()+body.String())
	} else {
		tg.docMessage(u.Params.TelegramID, "log.txt", header.String(), e.RW)
	}

	// For SMTP send log.eml for better preview.
	if e.Protocol.Category() == models.ProtoCategorySMTP {
		tg.docMessage(u.Params.TelegramID, "log.eml", header.String(), e.RW)
	}

	return nil
}
