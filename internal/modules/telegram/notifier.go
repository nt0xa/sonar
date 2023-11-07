package telegram

import (
	"unicode/utf8"

	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/modules"
)

const maxMessageSize = 4096

func (tg *Telegram) Notify(n *modules.Notification) error {
	header, body, err := tg.tmpl.RenderNotification(n)
	if err != nil {
		return err
	}

	if len(header+body) < maxMessageSize && utf8.ValidString(body) {
		tg.htmlMessage(n.User.Params.TelegramID, nil, header+"\n"+body)
	} else {
		tg.docMessage(n.User.Params.TelegramID, "log.txt", header, n.Event.RW)
	}

	// For SMTP send log.eml for better preview.
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP {
		tg.docMessage(n.User.Params.TelegramID, "log.eml", header, n.Event.RW)
	}

	return nil
}
