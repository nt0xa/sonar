package telegram

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"time"
	"unicode/utf8"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"golang.org/x/time/rate"
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

	rw := bytes.Join(n.Event.Data, nil)

	header, body, err := tg.tmpl.RenderNotification(n)
	if err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	if len(header+body) < maxMessageSize && utf8.ValidString(body) {
		tg.htmlMessage(ctx, chatID, nil, header+body)
	} else {
		tg.docMessage(ctx, chatID, "log.txt", header, rw)
	}

	// For SMTP send log.eml for better preview.
	if database.ProtoToCategory(n.Event.Protocol) == database.ProtoCategorySMTP && n.Event.Meta.SMTP != nil {
		data := n.Event.Meta.SMTP.Session.Data
		if data != "" {
			tg.docMessage(ctx, chatID, "log.eml", header, []byte(data))
			tg.docMessage(ctx, chatID, "log.txt", header, rw)
		}
	}

	return nil
}

func (tg *Telegram) NotifyBatch(ctx context.Context, ns []*modules.Notification) error {
	if ns[0].User.TelegramID == nil {
		return fmt.Errorf("user %d has no telegram id", ns[0].User.ID)
	}

	msg, err := tg.tmpl.RenderNotificationBatch(ns)
	if err != nil {
		return fmt.Errorf("telegram: %w", err)
	}

	tg.htmlMessage(ctx, *ns[0].User.TelegramID, nil, fmt.Sprintf(
		"#%s 📦 <b>%d more events</b>\n%s",
		html.EscapeString(ns[0].Payload.Name), len(ns), msg,
	))

	return nil
}

// RateLimit follows https://core.telegram.org/bots/faq: at most one message per second in a chat.
func (tg *Telegram) RateLimit() (rate.Limit, int) {
	return rate.Every(time.Second), 1
}
