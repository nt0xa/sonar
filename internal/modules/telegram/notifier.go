package telegram

import (
	"bytes"
	"fmt"
	"html/template"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

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

	var message tgbotapi.Chattable

	if len(header.String()+body.String()) < maxMessageSize && utf8.ValidString(body.String()) {
		// Send as message.
		msg := tgbotapi.NewMessage(u.Params.TelegramID, header.String()+body.String())
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
		message = msg
	} else {
		// Send as document.
		doc := tgbotapi.FileBytes{
			Name:  "log.txt",
			Bytes: e.RawData,
		}

		msg := tgbotapi.NewDocumentUpload(u.Params.TelegramID, doc)
		msg.Caption = header.String()
		msg.ParseMode = tgbotapi.ModeHTML
		message = msg
	}

	if _, err := tg.api.Send(message); err != nil {
		return fmt.Errorf("send message error: %w", err)
	}

	return nil
}
