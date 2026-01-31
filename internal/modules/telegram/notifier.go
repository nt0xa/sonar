package telegram

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
)

const maxMessageSize = 4096

func (tg *Telegram) Name() string {
	return "telegram"
}

func (tg *Telegram) Notify(ctx context.Context, n *modules.Notification) error {
	if n.User.TelegramID == nil {
		return fmt.Errorf("user %d has no telegram id", n.User.ID)
	}

	chatID := *n.User.TelegramID

	header, body, err := tg.tmpl.RenderNotification(n)
	if err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	if len(header+body) < maxMessageSize && utf8.ValidString(body) {
		tg.htmlMessage(ctx, chatID, nil, header+body)
	} else {
		tg.docMessage(ctx, chatID, "log.txt", header, n.Event.RW)
	}

	// For SMTP send log.eml for better preview.
	if database.ProtoToCategory(n.Event.Protocol) == database.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		data := n.Event.Meta.SMTP.Session.Data
		if data != "" {
			tg.docMessage(ctx, chatID, "log.eml", header, []byte(data))
			tg.docMessage(ctx, chatID, "log.txt", header, n.Event.RW)
		}
	}

	return nil
}
