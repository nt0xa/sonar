package telegram

import (
	"bytes"
	"fmt"
	"html/template"
	"unicode/utf8"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
)

const maxMessageSize = 4096

var (
	messageHeaderTemplate = template.Must(template.New("msg").Parse(`
<b>[{{ .Name }}]</b> {{ .Protocol }} from <code>{{ .RemoteAddr }}</code> at {{ .ReceivedAt }}
`))
	messageBodyTemplate = template.Must(template.New("msg").Parse(`
<pre>
{{ .Data }}
</pre>
`))
)

func (tg *Telegram) Notify(e *models.Event, u *models.User, p *models.Payload) error {

	var header, body bytes.Buffer

	headerData := struct {
		Name       string
		Protocol   string
		RemoteAddr string
		ReceivedAt string
	}{
		p.Name,
		e.Protocol,
		e.RemoteAddr.String(),
		e.ReceivedAt.Format("15:04:05 01.01.2006"),
	}

	if err := messageHeaderTemplate.Execute(&header, headerData); err != nil {
		return fmt.Errorf("message header render error: %w", err)
	}

	var data string

	if e.Protocol == "DNS" {
		data = utils.HexDump(e.RawData)
	} else {
		data = e.Data
	}

	bodyData := struct {
		Data string
	}{
		data,
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
