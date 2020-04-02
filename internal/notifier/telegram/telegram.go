package telegram

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/notifier"
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

type Notifier struct {
	tg *tgbotapi.BotAPI
}

// Check that Notifier implements notifier.Notifier interface
var _ notifier.Notifier = &Notifier{}

func New(cfg *Config) (*Notifier, error) {
	client := http.DefaultClient

	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

	}

	tg, err := tgbotapi.NewBotAPIWithClient(cfg.Token, client)
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	return &Notifier{tg}, nil
}

func (n *Notifier) Notify(e *notifier.Event, u *database.User, p *database.Payload) error {

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
		return fmt.Errorf("telegram message header render error: %w", err)
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
		return fmt.Errorf("telegram message body render error: %w", err)
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

	if _, err := n.tg.Send(message); err != nil {
		return fmt.Errorf("telegram message send error: %w", err)
	}

	return nil
}
