package telegram

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"github.com/Masterminds/sprig"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/russtone/sonar/internal/actions"
	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/cmd"
	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/results"
	"github.com/russtone/sonar/internal/utils"
	"github.com/russtone/sonar/internal/utils/errors"
)

type Telegram struct {
	api     *tgbotapi.BotAPI
	db      *database.DB
	cfg     *Config
	cmd     *cmd.Command
	actions actions.Actions
	bot     tgbotapi.User

	domain string
}

func New(cfg *Config, db *database.DB, actions actions.Actions, domain string) (*Telegram, error) {
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

	api, err := tgbotapi.NewBotAPIWithClient(cfg.Token, client)
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	bot, err := api.GetMe()
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	tg := &Telegram{
		api:     api,
		db:      db,
		cfg:     cfg,
		domain:  domain,
		bot:     bot,
		actions: actions,
	}

	tg.cmd = cmd.New(
		actions,
		&results.Text{
			Templates: results.DefaultTemplates(results.TemplateOptions{
				Markup: map[string]string{
					"<bold>":   "<b>",
					"</bold>":  "</b>",
					"<code>":   "<code>",
					"</code>":  "</code>",
					"<error>":  "",
					"</error>": "",
					"<pre>":    "<pre>",
					"</pre>":   "</pre>",
				},
				ExtraFuncs: template.FuncMap{
					"domain": func() string {
						return domain
					},
				},
				HTML: true,
			}),
			OnText: func(ctx context.Context, id, message string) {
				chatID, err := GetChatID(ctx)
				if err != nil {
					// TODO: logging
					return
				}
				tg.htmlMessage(chatID, message)
			},
		},
		cmd.PreExec(cmd.DefaultMessengersPreExec),
	)

	return tg, nil
}

func (tg *Telegram) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tg.api.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		chat := msg.Chat

		// Ignore error because user=nil is unauthorized user and there are
		// some commands available for unauthorized users (e.g. "/id")
		chatUser, _ := tg.db.UsersGetByParam(models.UserTelegramID, chat.ID)
		fromUser, _ := tg.db.UsersGetByParam(models.UserTelegramID, msg.From.ID)

		// Create user for group on group creation or when bot added to already
		// existing group.
		if tg.isAddedToGroup(msg) && fromUser != nil {
			rnd, _ := utils.GenerateRandomString(4)

			u := &models.User{
				Name:      fmt.Sprintf("shared-%s", rnd),
				CreatedBy: &fromUser.ID,
				Params: models.UserParams{
					TelegramID: chat.ID,
				},
			}

			if err := tg.db.UsersCreate(u); err != nil {
				tg.handleError(chat.ID, errors.Internal(err))
			}

			continue
		}

		// Delete group user when it is removed from group.
		if chat.IsGroup() && tg.isDeletedFromGroup(msg) && chatUser != nil {
			if err := tg.db.UsersDelete(chatUser.ID); err != nil {
				tg.handleError(chat.ID, errors.Internal(err))
			}

			continue
		}

		ctx := SetChatID(context.Background(), chat.ID)

		// Only registered users should be able to use bot.
		if chat.IsGroup() && fromUser == nil {
			ctx = actionsdb.SetUser(ctx, nil)
		} else {
			ctx = actionsdb.SetUser(ctx, chatUser)
		}

		tg.cmd.ParseAndExec(ctx, update.Message.Text)
	}

	return nil
}

func (tg *Telegram) isAddedToGroup(msg *tgbotapi.Message) bool {
	if msg.GroupChatCreated {
		return true
	}

	if msg.NewChatMembers != nil {
		for _, m := range *msg.NewChatMembers {
			if tg.bot.ID == m.ID {
				return true
			}
		}
	}

	return false
}

func (tg *Telegram) isDeletedFromGroup(msg *tgbotapi.Message) bool {
	return msg.LeftChatMember != nil && msg.LeftChatMember.ID == tg.bot.ID
}

// TODO
// func (tg *Telegram) preExec(root *cobra.Command, u *actions.User) {
// 	cmd := &cobra.Command{
// 		Use:   "id",
// 		Short: "Show current telegram chat id",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			chatID, err := GetChatID(cmd.Context())
// 			if err != nil {
// 				return errors.Internal(err)
// 			}
//
// 			tg.txtMessage(chatID, fmt.Sprintf("%d", chatID))
// 		},
// 	}
//
// 	root.AddCommand(cmd)
// }

func (tg *Telegram) handleError(chatID int64, err errors.Error) {
	tg.txtMessage(chatID, err.Error())
}

func (tg *Telegram) txtMessage(chatID int64, txt string) {
	var tpl bytes.Buffer

	if err := codeTemplate.Execute(&tpl, txt); err != nil {
		return
	}

	tg.htmlMessage(chatID, tpl.String())
}

func (tg *Telegram) htmlMessage(chatID int64, html string) {
	msg := tgbotapi.NewMessage(
		chatID,
		html,
	)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	_, err := tg.api.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
}

func (tg *Telegram) docMessage(chatID int64, name string, caption string, data []byte) {
	doc := tgbotapi.FileBytes{
		Name:  name,
		Bytes: data,
	}

	msg := tgbotapi.NewDocumentUpload(chatID, doc)
	msg.Caption = caption
	msg.ParseMode = tgbotapi.ModeHTML
	tg.api.Send(msg)
}

func tpl(s string) *template.Template {
	return template.Must(template.
		New("").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{
			// This is nesessary for templates to compile.
			// It will be replaced later with correct function.
			"domain": func() string { return "" },
		}).
		Parse(s),
	)
}

func (tg *Telegram) getDomain() string {
	return tg.domain
}

func (tg *Telegram) txtResult(ctx context.Context, txt string) {
	u, err := actionsdb.GetUser(ctx)
	if err != nil {
		return
	}

	tg.txtMessage(u.Params.TelegramID, txt)
}

func (tg *Telegram) tplResult(ctx context.Context, tpl *template.Template, data interface{}) {
	u, err := actionsdb.GetUser(ctx)
	if err != nil {
		return
	}

	tpl.Funcs(template.FuncMap{
		"domain": tg.getDomain,
	})

	buf := &bytes.Buffer{}

	if err := tpl.Execute(buf, data); err != nil {
		tg.handleError(u.Params.TelegramID, errors.Internal(err))
	}

	tg.htmlMessage(u.Params.TelegramID, buf.String())
}

var codeTemplate = tpl(`<code>{{ . }}</code>`)
