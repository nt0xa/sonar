package slack

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/templates"
	"github.com/nt0xa/sonar/pkg/telemetry"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	client  *slack.Client
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
) (*Slack, error) {
	client := slack.New(
		cfg.BotToken,
		slack.OptionAppLevelToken(cfg.AppToken),
	)

	if _, err := client.AuthTest(); err != nil {
		return nil, fmt.Errorf("slack auth test failed: %w", err)
	}

	tmpl := templates.New(domain,
		templates.Default(
			templates.HTMLEscape(false),
			templates.Markup(
				templates.Bold("*", "*"),
				templates.CodeBlock("```", "```"),
				templates.CodeInline("`", "`"),
			),
		),
	)

	s := &Slack{
		client:  client,
		db:      db,
		tel:     tel,
		log:     log,
		cmd:     cmd.New(actions),
		actions: actions,
		tmpl:    tmpl,
		domain:  domain,
	}

	return s, nil
}

func (s *Slack) Start() error {
	socketClient := socketmode.New(s.client)

	go func() {
		for evt := range socketClient.Events {
			switch evt.Type {
			case socketmode.EventTypeSlashCommand:
				cmd, _ := evt.Data.(slack.SlashCommand)

				socketClient.Ack(*evt.Request, map[string]any{
					"text": s.processCommand(context.TODO(), cmd),
				})
			}
		}
	}()

	s.log.Info("Starting Slack socket client")
	if err := socketClient.Run(); err != nil {
		s.log.Error("Socket client error", "err", err)
		return err
	}

	return nil
}

func (s *Slack) processCommand(ctx context.Context, cmd slack.SlashCommand) *string {
	chatUser, _ := s.db.UsersGetByParam(ctx, models.UserSlackID, cmd.UserID)
	ctx = actionsdb.SetUser(ctx, chatUser)

	reply := ""

	stdout, stderr, err := s.cmd.ParseAndExec(ctx, cmd.Command+" "+cmd.Text,
		func(ctx context.Context, res actions.Result) error {
			content, err := s.tmpl.RenderResult(res)
			if err != nil {
				return err
			}
			reply = content
			return nil
		},
	)

	if err != nil {
		reply = err.Error()
	}

	if stdout != "" {
		reply = stdout
	}

	if stderr != "" {
		reply = stderr
	}

	return &reply
}
