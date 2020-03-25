package telegram

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/controller"
	"github.com/bi-zone/sonar/internal/database"
)

var (
	helpMessage = "" +
		"`/new <name>` - create new payload\n" +
		"`/del <name>` - delete payload\n" +
		"`/list <substr>` - list your payloads which contains `<substr>`\n"
	listPayloadTemplate = template.Must(template.New("msg").
				Parse(`<b>[{{ .Name }}]</b> - <code>{{ .Subdomain }}.{{ .Domain }}</code>`))
)

type Bot struct {
	tg     *tgbotapi.BotAPI
	db     *database.DB
	domain string
	admin  int64
}

var _ controller.Controller = &Bot{}

func New(cfg *Config, db *database.DB, domain string) (controller.Controller, error) {
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

	// Admin
	if cfg.Admin != 0 {
		u, err := db.UsersGetByName("admin")

		if err != nil {
			return nil, fmt.Errorf("telegram: fail to get admin user from db: %w", err)
		}

		u.Params.TelegramID = cfg.Admin

		if err := db.UsersUpdate(u); err != nil {
			return nil, fmt.Errorf("telegram: fail to set admin id in db: %w", err)
		}
	}

	tg, err := tgbotapi.NewBotAPIWithClient(cfg.Token, client)
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	return &Bot{tg, db, domain, cfg.Admin}, nil
}

func (b *Bot) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.tg.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		var err error

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		cmd := update.Message.Text
		chatID := update.Message.Chat.ID

		switch {
		case strings.HasPrefix(cmd, "/help"):
			err = b.checkUser(chatID, cmd, b.helpCmd)

		case strings.HasPrefix(cmd, "/new"):
			err = b.checkUser(chatID, cmd, b.newCmd)

		case strings.HasPrefix(cmd, "/del"):
			err = b.checkUser(chatID, cmd, b.delCmd)

		case strings.HasPrefix(cmd, "/list"):
			err = b.checkUser(chatID, cmd, b.listCmd)

		// Admin commans
		case strings.HasPrefix(cmd, "/useradd"):
			err = b.checkAdmin(chatID, cmd, b.userAddCmd)

		case strings.HasPrefix(cmd, "/userdel"):
			err = b.checkAdmin(chatID, cmd, b.userDelCmd)
		}

		if e, ok := err.(*Error); ok {
			_, _ = b.tg.Send(tgbotapi.NewMessage(
				chatID,
				e.Msg,
			))
			continue
		} else if err != nil {
			log.Println(err)
		}

	}

	return nil
}

type cmdFunc func(chatID int64, cmd string, u *database.User) error

func (b *Bot) checkUser(chatID int64, cmd string, next cmdFunc) error {
	u, err := b.db.UsersGetByParams(&database.UserParams{TelegramID: chatID})

	if err != nil {
		return ErrUnauthorizedAccess
	}

	return next(chatID, cmd, u)
}

func (b *Bot) checkAdmin(chatID int64, cmd string, next cmdFunc) error {
	if chatID != b.admin {
		return nil
	}

	u, err := b.db.UsersGetByParams(&database.UserParams{TelegramID: chatID})

	if err != nil {
		return ErrUnauthorizedAccess
	}

	return next(chatID, cmd, u)
}

func (b *Bot) helpCmd(chatID int64, cmd string, u *database.User) error {
	msg := tgbotapi.NewMessage(chatID, helpMessage)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.tg.Send(msg); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (b *Bot) newCmd(chatID int64, cmd string, u *database.User) error {
	subdomain, err := GenerateRandomString(4)
	if err != nil {
		return ErrInternal.SetError(err)
	}

	args := strings.SplitN(cmd, " ", 2)

	if len(args) != 2 || args[1] == "" {
		return &Error{Msg: `Argument "name" is required`}
	}

	name := args[1]

	if _, err := b.db.PayloadsGetByUserAndName(u.ID, name); err != sql.ErrNoRows {
		return &Error{Msg: fmt.Sprintf("You already have payload with name %q", name)}
	}

	p := &database.Payload{
		UserID:    u.ID,
		Subdomain: subdomain,
		Name:      name,
	}

	if err := b.db.PayloadsCreate(p); err != nil {
		return ErrInternal.SetError(err)
	}

	text := fmt.Sprintf("`%s.%s`", p.Subdomain, b.domain)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	if _, err := b.tg.Send(msg); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (b *Bot) delCmd(chatID int64, cmd string, u *database.User) error {
	args := strings.SplitN(cmd, " ", 2)

	if len(args) < 2 || args[1] == "" {
		return &Error{Msg: `Argument "name" is required`}
	}

	name := args[1]

	p, err := b.db.PayloadsGetByUserAndName(u.ID, name)
	if err == sql.ErrNoRows {
		return &Error{Msg: "Payload not found"}
	} else if err != nil {
		return ErrInternal.SetError(err)
	}

	if err := b.db.PayloadsDelete(p.ID); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := b.tg.Send(tgbotapi.NewMessage(
		chatID,
		"Payload deleted",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (b *Bot) listCmd(chatID int64, cmd string, u *database.User) error {
	name := ""
	args := strings.SplitN(cmd, " ", 2)

	if len(args) == 2 {
		name = args[1]
	}

	pp, err := b.db.PayloadsFindByUserAndName(u.ID, name)
	if err != nil {
		return ErrInternal.SetError(err)
	}

	if len(pp) == 0 {
		return &Error{Msg: "You don't have any payloads yet"}
	}

	text := ""

	for _, p := range pp {
		var tpl bytes.Buffer

		data := struct {
			Name      string
			Subdomain string
			Domain    string
		}{
			p.Name,
			p.Subdomain,
			b.domain,
		}

		if err := listPayloadTemplate.Execute(&tpl, data); err != nil {
			return err
		}
		text += tpl.String() + "\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true

	if _, err := b.tg.Send(msg); err != nil {
		return err
	}

	return nil
}

func (b *Bot) userAddCmd(chatID int64, cmd string, adm *database.User) error {
	args := strings.Split(cmd, " ")
	if len(args) != 3 {
		return &Error{Msg: "Invalid arguments count"}
	}

	var (
		id  int64
		err error
	)

	id, err = strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &Error{Msg: "Invalid user id"}
	}

	u := &database.User{
		Name: args[2],
		Params: database.UserParams{
			TelegramID: id,
		},
	}
	if err := b.db.UsersCreate(u); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := b.tg.Send(tgbotapi.NewMessage(
		chatID,
		"user created",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (b *Bot) userDelCmd(chatID int64, cmd string, adm *database.User) error {
	args := strings.Split(cmd, " ")
	if len(args) != 2 {
		return &Error{Msg: "Invalid arguments count"}
	}

	var (
		id  int64
		err error
	)

	id, err = strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &Error{Msg: "Invalid user id"}
	}

	u, err := b.db.UsersGetByParams(&database.UserParams{TelegramID: id})
	if err != nil {
		return ErrNotFound
	}

	if err := b.db.UsersDelete(u.ID); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := b.tg.Send(tgbotapi.NewMessage(
		chatID,
		"user deleted",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
