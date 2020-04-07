package telegram

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils"
)

var (
	helpMessage = "" +
		"`/new <name>` - create new payload\n" +
		"`/del <name>` - delete payload\n" +
		"`/list <substr>` - list your payloads which contains `<substr>`\n" +
		"`/me` - user info\n"

	listPayloadTemplate = template.Must(template.New("msg").
				Parse("<b>[{{ .Name }}]</b> - <code>{{ .Subdomain }}.{{ .Domain }}</code>"))

	meTemplate = template.Must(template.New("msg").
			Parse("" +
			"<b>Telegram ID:</b> <code>{{ .TelegramID }}</code>\n" +
			"<b>API token:</b> <code>{{ .APIToken }}</code>",
		))
)

func (tg *Telegram) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tg.api.GetUpdatesChan(u)
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
			err = tg.checkUser(chatID, cmd, tg.helpCmd)

		case strings.HasPrefix(cmd, "/new"):
			err = tg.checkUser(chatID, cmd, tg.newCmd)

		case strings.HasPrefix(cmd, "/del"):
			err = tg.checkUser(chatID, cmd, tg.delCmd)

		case strings.HasPrefix(cmd, "/list"):
			err = tg.checkUser(chatID, cmd, tg.listCmd)

		case strings.HasPrefix(cmd, "/me"):
			err = tg.checkUser(chatID, cmd, tg.meCmd)

		// Admin commans
		case strings.HasPrefix(cmd, "/useradd"):
			err = tg.checkAdmin(chatID, cmd, tg.userAddCmd)

		case strings.HasPrefix(cmd, "/userdel"):
			err = tg.checkAdmin(chatID, cmd, tg.userDelCmd)
		}

		if e, ok := err.(*Error); ok {
			_, _ = tg.api.Send(tgbotapi.NewMessage(
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

func (tg *Telegram) checkUser(chatID int64, cmd string, next cmdFunc) error {
	u, err := tg.db.UsersGetByParams(&database.UserParams{TelegramID: chatID})

	if err != nil {
		return ErrUnauthorizedAccess
	}

	return next(chatID, cmd, u)
}

func (tg *Telegram) checkAdmin(chatID int64, cmd string, next cmdFunc) error {
	if chatID != tg.cfg.Admin {
		return nil
	}

	u, err := tg.db.UsersGetByParams(&database.UserParams{TelegramID: chatID})

	if err != nil {
		return ErrUnauthorizedAccess
	}

	return next(chatID, cmd, u)
}

func (tg *Telegram) helpCmd(chatID int64, cmd string, u *database.User) error {
	msg := tgbotapi.NewMessage(chatID, helpMessage)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := tg.api.Send(msg); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (tg *Telegram) newCmd(chatID int64, cmd string, u *database.User) error {
	subdomain, err := utils.GenerateRandomString(4)
	if err != nil {
		return ErrInternal.SetError(err)
	}

	args := strings.SplitN(cmd, " ", 2)

	if len(args) != 2 || args[1] == "" {
		return &Error{Msg: `Argument "name" is required`}
	}

	name := args[1]

	if _, err := tg.db.PayloadsGetByUserAndName(u.ID, name); err != sql.ErrNoRows {
		return &Error{Msg: fmt.Sprintf("You already have payload with name %q", name)}
	}

	p := &database.Payload{
		UserID:    u.ID,
		Subdomain: subdomain,
		Name:      name,
	}

	if err := tg.db.PayloadsCreate(p); err != nil {
		return ErrInternal.SetError(err)
	}

	text := fmt.Sprintf("`%s.%s`", p.Subdomain, tg.domain)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true

	if _, err := tg.api.Send(msg); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (tg *Telegram) delCmd(chatID int64, cmd string, u *database.User) error {
	args := strings.SplitN(cmd, " ", 2)

	if len(args) < 2 || args[1] == "" {
		return &Error{Msg: `Argument "name" is required`}
	}

	name := args[1]

	p, err := tg.db.PayloadsGetByUserAndName(u.ID, name)
	if err == sql.ErrNoRows {
		return &Error{Msg: "Payload not found"}
	} else if err != nil {
		return ErrInternal.SetError(err)
	}

	if err := tg.db.PayloadsDelete(p.ID); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := tg.api.Send(tgbotapi.NewMessage(
		chatID,
		"Payload deleted",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (tg *Telegram) listCmd(chatID int64, cmd string, u *database.User) error {
	name := ""
	args := strings.SplitN(cmd, " ", 2)

	if len(args) == 2 {
		name = args[1]
	}

	pp, err := tg.db.PayloadsFindByUserAndName(u.ID, name)
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
			tg.domain,
		}

		if err := listPayloadTemplate.Execute(&tpl, data); err != nil {
			return err
		}
		text += tpl.String() + "\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true

	if _, err := tg.api.Send(msg); err != nil {
		return err
	}

	return nil
}

func (tg *Telegram) meCmd(chatID int64, cmd string, u *database.User) error {
	var tpl bytes.Buffer

	if err := meTemplate.Execute(&tpl, u.Params); err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatID, strings.TrimPrefix(tpl.String(), "\n"))
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err := tg.api.Send(msg); err != nil {
		return err
	}

	return nil
}

func (tg *Telegram) userAddCmd(chatID int64, cmd string, adm *database.User) error {
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
	if err := tg.db.UsersCreate(u); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := tg.api.Send(tgbotapi.NewMessage(
		chatID,
		"user created",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}

func (tg *Telegram) userDelCmd(chatID int64, cmd string, adm *database.User) error {
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

	u, err := tg.db.UsersGetByParams(&database.UserParams{TelegramID: id})
	if err != nil {
		return ErrNotFound
	}

	if err := tg.db.UsersDelete(u.ID); err != nil {
		return ErrInternal.SetError(err)
	}

	if _, err := tg.api.Send(tgbotapi.NewMessage(
		chatID,
		"user deleted",
	)); err != nil {
		return ErrInternal.SetError(err)
	}

	return nil
}
