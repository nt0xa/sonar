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

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/templates"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type Telegram struct {
	api     *tgbotapi.BotAPI
	db      *database.DB
	tel     telemetry.Telemetry
	log     *slog.Logger
	cmd     *cmd.Command
	actions actions.Actions
	tmpl    *templates.Templates

	domain string
}

func New(
	cfg *Config,
	db *database.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	actions actions.Actions,
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
				templates.CodeBlock(
					`<pre><code class="language-{{ codeLanguage $protocol }}">`,
					"</code></pre>",
				),
			),
			templates.ExtraFunc("codeLanguage", func(proto string) string {
				// https://github.com/TelegramMessenger/libprisma#supported-languages
				category := models.Proto{Name: proto}.Category()

				switch category {
				case models.ProtoCategoryHTTP:
					return "http"
				case models.ProtoCategoryDNS:
					return "dns-zone"
				case models.ProtoCategorySMTP:
					return "markdown"
				case models.ProtoCategoryFTP:
					return "log"
				}
				return ""
			}),
		),
	)

	tg := &Telegram{
		api:     api,
		db:      db,
		log:     log,
		tel:     tel,
		domain:  domain,
		actions: actions,
		tmpl:    tmpl,
	}

	tg.cmd = cmd.New(
		actions,
		cmd.PreExec(tg.preExec),
	)

	return tg, nil
}

func (tg *Telegram) preExec(acts *actions.Actions, root *cobra.Command) {
	cmd.DefaultMessengersPreExec(acts, root)

	c := &cobra.Command{
		Use:   "id",
		Short: "Show current telegram chat id",
		RunE: func(cmd *cobra.Command, args []string) error {
			mi, err := getMsgInfo(cmd.Context())
			if err != nil {
				return errors.Internal(err)
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

	// Ignore error because user=nil is unauthorized user and there are
	// some commands available for unauthorized users (e.g. "/id")
	chatUser, _ := tg.db.UsersGetByParam(ctx, models.UserTelegramID, chat.ID)
	ctx = actionsdb.SetUser(ctx, chatUser)
	ctx = setMsgInfo(ctx, chat.ID, msg.MessageID)

	stdout, stderr, err := tg.cmd.ParseAndExec(ctx, msg.Text,
		func(ctx context.Context, res actions.Result) error {
			s, err := tg.tmpl.RenderResult(res)
			if err != nil {
				return err
			}
			tg.htmlMessage(ctx, chat.ID, &msg.MessageID, s)
			return nil
		},
	)

	if err != nil {
		tg.handleError(ctx, chat.ID, &msg.MessageID, err)
		return nil
	}

	if stdout != "" {
		tg.htmlMessage(ctx, chat.ID, &msg.MessageID, stdout)
	}

	if stderr != "" {
		tg.htmlMessage(ctx, chat.ID, &msg.MessageID, stderr)
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
