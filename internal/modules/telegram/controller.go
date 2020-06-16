package telegram

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/cmd"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

var (
	helpTemplate = `<code>{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}</code>`

	usageTemplate = `<code>
Usage:{{if .Runnable}}{{if .HasParent}}
  {{.UseLine | replace "sonarctl " "/"}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
	{{if .HasParent}}{{.CommandPath | replace "sonarctl " "/"}} {{else}}/{{end}}[command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  /{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{if .HasParent}}{{.CommandPath | replace "sonarctl " "/"}} {{else}}/{{end}}[command] --help" for more information about a command.{{end}}
</code>`

	codeTemplate = tpl(`<code>{{ . }}</code>`)

	listPayloadTemplate = tpl(`{{range .Payloads}}<b>[{{ .Name }}]</b> - <code>{{ .Subdomain }}.{{ $.Domain }}</code>
{{else}}you don't have any payloads yet{{end}}`)

	meTemplate = tpl("" +
		"<b>Telegram ID:</b> <code>{{ .TelegramID }}</code>\n" +
		"<b>API token:</b> <code>{{ .APIToken }}</code>",
	)
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

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		u, err := tg.db.UsersGetByParams(&models.UserParams{TelegramID: chatID})
		if err != nil {
			tg.handleError(chatID, errors.Unauthorized())
			continue
		}

		if out, err := tg.handleCommand(u, text); err != nil {
			tg.handleError(chatID, err)
		} else if out != "" {
			tg.htmlMessage(chatID, out)
		}
	}

	return nil
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

func (tg *Telegram) handleCommand(u *models.User, text string) (string, errors.Error) {
	// Prepare context
	ctx := context.Background()
	ctx = cmd.SetUser(ctx, u)

	// Create root command
	root := cmd.RootCmd(tg.actions, tg.handleResult)

	// Set args
	args := strings.Split(strings.TrimLeft(text, "/"), " ")
	root.SetArgs(args)

	// Set templates
	root.SetHelpTemplate(helpTemplate)
	root.SetUsageTemplate(usageTemplate)

	// Set output
	bb := &bytes.Buffer{}
	root.SetErr(bb)
	root.SetOut(bb)

	// Execute command with context
	if err := root.ExecuteContext(ctx); err != nil {
		e, ok := err.(errors.Error)
		if !ok {
			return "", errors.Internal(err)
		}

		return "", e
	}

	return bb.String(), nil
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
