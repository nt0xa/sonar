package telegram

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/modules"
)

const maxMessageSize = 4096

func (tg *Telegram) Name() string {
	return "telegram"
}

func (tg *Telegram) Notify(ctx context.Context, n *modules.Notification) error {
	header, body, err := tg.tmpl.RenderNotification(n)
	if err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	if len(header+body) < maxMessageSize && utf8.ValidString(body) {
		tg.htmlMessage(ctx, n.User.Params.TelegramID, nil, header+body)
	} else {
		tg.docMessage(ctx, n.User.Params.TelegramID, "log.txt", header, n.Event.RW)
	}

	// For SMTP send log.eml for better preview.
	if n.Event.Protocol.Category() == models.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		data := n.Event.Meta.SMTP.Session.Data
		if data != "" {
			tg.docMessage(ctx, n.User.Params.TelegramID, "log.eml", header, []byte(data))
			tg.docMessage(ctx, n.User.Params.TelegramID, "log.txt", header, n.Event.RW)
		}
	}

	return nil
}
