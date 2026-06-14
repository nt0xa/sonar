package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/templates"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type Telegram struct {
	api  *tgbotapi.BotAPI
	tel  telemetry.Telemetry
	log  *slog.Logger
	cmd  *cmd.Command
	svc  service.ServerService
	tmpl *templates.Templates

	domain string
}

func New(
	cfg *Config,
	log *slog.Logger,
	tel telemetry.Telemetry,
	svc service.ServerService,
	domain string,
) (*Telegram, error) {
	client := http.DefaultClient

	// Proxy
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("telegram invalid proxy url: %w", err)
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

	}

	api, err := tgbotapi.NewBotAPIWithClient(cfg.Token, tgbotapi.APIEndpoint, client)
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	tmpl := templates.New(domain,
		templates.Default(
			templates.Markup(
				templates.Bold("<b>", "</b>"),
				templates.CodeInline("<code>", "</code>"),
			),
			templates.ExtraFunc("codeLanguage", func(proto string) string {
				// https://github.com/TelegramMessenger/libprisma#supported-languages
				category := database.ProtoToCategory(proto)

				switch category {
				case database.ProtoCategoryHTTP:
					return "http"
				case database.ProtoCategoryDNS:
					return "dns-zone"
				case database.ProtoCategorySMTP:
					return "markdown"
				case database.ProtoCategoryFTP:
					return "log"
				}
				return ""
			}),
		),
		templates.PerTemplate(templates.NotificationBodyID,
			templates.Markup(
				templates.CodeBlock(
					`<pre><code class="language-{{ codeLanguage .Event.Protocol }}">`,
					"</code></pre>",
				),
			),
		),
	)

	tg := &Telegram{
		api:    api,
		log:    log,
		tel:    tel,
		domain: domain,
		svc:    svc,
		tmpl:   tmpl,
	}

	tg.cmd = cmd.New(
		svc,
		cmd.PreExec(tg.preExec),
	)

	return tg, nil
}

func (tg *Telegram) preExec(root *cobra.Command) {
	cmd.DefaultMessengersPreExec(root)

	c := &cobra.Command{
		Use:   "id",
		Short: "Show current telegram chat id",
		RunE: func(cmd *cobra.Command, args []string) error {
			mi, err := getMsgInfo(cmd.Context())
			if err != nil {
				return err
			}

			tg.htmlMessage(cmd.Context(), mi.chatID, &mi.msgID, fmt.Sprintf("<code>%d</code>", mi.chatID))
			return nil
		},
	}

	root.AddCommand(c)
}

func (tg *Telegram) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.api.GetUpdatesChan(u)

	for update := range updates {
		ctx := context.Background()

		if err := tg.processUpdateWithTelemetry(ctx, update); err != nil {
			tg.log.Error(
				"Update processing error",
				"err", err,
				"update", update,
			)
		}
	}

	return nil
}

func (tg *Telegram) processUpdateWithTelemetry(ctx context.Context, update tgbotapi.Update) error {
	var msg *tgbotapi.Message

	switch {
	case update.Message != nil:
		msg = update.Message
	case update.EditedMessage != nil:
		msg = update.EditedMessage

	default:
		return fmt.Errorf("telegram update has no message: %v", update)
	}

	ctx, span := tg.tel.TraceStart(ctx, "telegram.update",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.Int("telegram.update.id", update.UpdateID),
			attribute.Int64("telegram.chat.id", msg.Chat.ID),
			attribute.String("telegram.chat.type", msg.Chat.Type),
		),
	)
	defer span.End()

	return tg.processUpdate(ctx, msg)
}

func (tg *Telegram) processUpdate(ctx context.Context, msg *tgbotapi.Message) error {
	chat := msg.Chat

	// Ignore the error: an unauthenticated user keeps the original context, so
	// commands available without auth (e.g. "/id") still work.
	if authCtx, err := tg.svc.AuthContextByTelegramID(ctx, chat.ID); err == nil {
		ctx = authCtx
	}
	ctx = setMsgInfo(ctx, chat.ID, msg.MessageID)

	res, err := tg.cmd.ParseAndExec(ctx, msg.Text)
	if err != nil {
		tg.handleError(ctx, chat.ID, &msg.MessageID, err)
		return nil
	}

	switch v := res.(type) {
	case string:
		// Help/usage text produced by cobra (no leaf result).
		if v != "" {
			tg.htmlMessage(ctx, chat.ID, &msg.MessageID, v)
		}
	default:
		s, err := tg.tmpl.RenderResult(v)
		if err != nil {
			tg.handleError(ctx, chat.ID, &msg.MessageID, err)
			return nil
		}
		tg.htmlMessage(ctx, chat.ID, &msg.MessageID, s)
	}

	return nil
}

func (tg *Telegram) handleError(ctx context.Context, chatID int64, msgID *int, err error) {
	tg.htmlMessage(ctx, chatID, msgID, err.Error())
}

func (tg *Telegram) htmlMessage(ctx context.Context, chatID int64, msgID *int, html string) {
	_, span := tg.tel.TraceStart(ctx, "telegram.htmlMessage",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	defer span.End()
	msg := tgbotapi.NewMessage(
		chatID,
		html,
	)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	if msgID != nil {
		msg.ReplyToMessageID = *msgID
	}
	_, err := tg.api.Send(msg)
	if err != nil {
		tg.log.Error("Send message error",
			"err", err,
		)
	}
}

func (tg *Telegram) docMessage(ctx context.Context, chatID int64, name string, caption string, data []byte) {
	_, span := tg.tel.TraceStart(ctx, "telegram.docMessage",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	doc := tgbotapi.FileBytes{
		Name:  name,
		Bytes: data,
	}
	msg := tgbotapi.NewDocument(chatID, doc)
	msg.Caption = caption
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := tg.api.Send(msg); err != nil {
		tg.log.Error("Send message error",
			"err", err,
		)
	}
}
