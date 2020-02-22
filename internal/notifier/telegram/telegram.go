package telegram

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/notifier"
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

func (n *Notifier) Notify(e *notifier.Event, p *database.Payload) error {

	var header, body bytes.Buffer

	data := struct {
		Name       string
		Protocol   string
		RemoteAddr string
		ReceivedAt string
		Data       string
	}{
		p.Name,
		e.Protocol,
		e.RemoteAddr.String(),
		e.ReceivedAt.Format("15:04:05 01.01.2006"),
		e.Data,
	}

	if err := messageHeaderTemplate.Execute(&header, data); err != nil {
		return fmt.Errorf("telegram message header render error: %w", err)
	}

	if err := messageBodyTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("telegram message body render error: %w", err)
	}

	var msg tgbotapi.MessageConfig

	if len(header.String()+body.String()) < maxMessageSize {
		// Send as message.
		msg = tgbotapi.NewMessage(p.UserID, header.String()+body.String())
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
	} else {
		// Send as document.
		rdr := tgbotapi.FileReader{
			Name:   "log.txt",
			Reader: bytes.NewReader([]byte(e.Data)),
			Size:   -1,
		}

		msg := tgbotapi.NewDocumentUpload(p.UserID, rdr)
		msg.Caption = header.String()
		msg.ParseMode = tgbotapi.ModeHTML
	}

	if _, err := n.tg.Send(msg); err != nil {
		return fmt.Errorf("telegram message send error: %w", err)
	}

	return nil
}
