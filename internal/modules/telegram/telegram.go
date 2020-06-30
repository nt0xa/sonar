package telegram

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/cobra"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/cmd"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/internal/utils/errors"
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

	cmd := &cmd.Command{
		Actions:       actions,
		ResultHandler: tg.handleResult,
		PreExec:       tg.preExec,
	}

	tg.cmd = cmd

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

		var user *models.User

		// Only registered users should be able to use bot.
		if chat.IsGroup() && fromUser == nil {
			user = nil
		} else {
			user = chatUser
		}

		ctx := SetChatID(context.Background(), chat.ID)
		args := strings.Split(strings.TrimLeft(update.Message.Text, "/"), " ")

		if out, err := tg.cmd.Exec(ctx, user, args); err != nil {
			tg.handleError(chat.ID, err)
		} else if out != "" {
			tg.htmlMessage(chat.ID, out)
		}
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

func (tg *Telegram) preExec(root *cobra.Command, user *models.User) {
	root.SetHelpTemplate(helpTemplate)
	root.SetUsageTemplate(usageTemplate)

	cmd := &cobra.Command{
		Use:   "id",
		Short: "Show current telegram chat id",
		RunE: cmd.RunE(func(cmd *cobra.Command, args []string) errors.Error {
			chatID, err := GetChatID(cmd.Context())
			if err != nil {
				return errors.Internal(err)
			}

			tg.txtMessage(chatID, fmt.Sprintf("%d", chatID))

			return nil
		}),
	}

	root.AddCommand(cmd)
}

func (tg *Telegram) handleResult(ctx context.Context, res interface{}) {
	var (
		tpl bytes.Buffer
		err error
	)

	u, err := cmd.GetUser(ctx)
	if err != nil {
		return
	}

	switch r := res.(type) {

	case actions.CreatePayloadResult:
		tg.txtMessage(u.Params.TelegramID, fmt.Sprintf("%s.%s", r.Subdomain, tg.domain))

	case actions.ListPayloadsResult:
		data := struct {
			Payloads actions.ListPayloadsResult
			Domain   string
		}{r, tg.domain}
		err = listPayloadTemplate.Execute(&tpl, data)
		tg.htmlMessage(u.Params.TelegramID, tpl.String())

	case actions.CreateUserResult:
		tg.txtMessage(u.Params.TelegramID, fmt.Sprintf("user %q created", r.Name))

	case *actions.MessageResult:
		tg.txtMessage(u.Params.TelegramID, r.Message)
	}

	if err != nil {
		tg.handleError(u.Params.TelegramID, errors.Internal(err))
		return
	}

}

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
	tg.api.Send(msg)
}
