package slack

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nt0xa/sonar/internal/cmd"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/internal/templates"
	"github.com/nt0xa/sonar/pkg/telemetry"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type Slack struct {
	client *slack.Client
	tel    telemetry.Telemetry
	log    *slog.Logger
	cmd    *cmd.Command
	svc    service.ServerService
	tmpl   *templates.Templates

	domain string
}

func New(
	cfg *Config,
	log *slog.Logger,
	tel telemetry.Telemetry,
	svc service.ServerService,
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
		client: client,
		tel:    tel,
		log:    log,
		cmd:    cmd.New(svc),
		svc:    svc,
		tmpl:   tmpl,
		domain: domain,
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
	// Ignore the error: an unauthenticated user keeps the original context and
	// service commands then fail with a clean "unauthorized".
	if authCtx, err := s.svc.AuthContextBySlackID(ctx, cmd.UserID); err == nil {
		ctx = authCtx
	}

	reply := ""

	res, err := s.cmd.ParseAndExec(ctx, cmd.Command+" "+cmd.Text)
	if err != nil {
		reply = err.Error()
		return &reply
	}

	switch v := res.(type) {
	case string:
		// Help/usage text produced by cobra (no leaf result).
		reply = v
	default:
		content, err := s.tmpl.RenderResult(v)
		if err != nil {
			reply = err.Error()
		} else {
			reply = content
		}
	}

	return &reply
}
