package server

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/modules/api"
	"github.com/nt0xa/sonar/internal/modules/lark"
	"github.com/nt0xa/sonar/internal/modules/slack"
	"github.com/nt0xa/sonar/internal/modules/telegram"
	"github.com/nt0xa/sonar/internal/service"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type Controller interface {
	Start() error
}

type ModulesConfig struct {
	Enabled []string

	// TODO: dynamic modules config (something like json.RawMessage) to be able to not include
	// unnecessary modules in binary.
	Telegram telegram.Config
	API      api.Config
	Lark     lark.Config
	Slack    slack.Config
}

func (c ModulesConfig) Validate() error {
	rules := make([]*validation.FieldRules, 0)
	rules = append(rules, validation.Field(&c.Enabled,
		validation.Each(validation.In("telegram", "api", "lark", "slack"))))

	// TODO: dynamic modules registration. Something like sql drivers
	for _, name := range c.Enabled {
		switch name {

		case "telegram":
			rules = append(rules, validation.Field(&c.Telegram))

		case "api":
			rules = append(rules, validation.Field(&c.API))

		case "lark":
			rules = append(rules, validation.Field(&c.Lark))
		}
	}

	return validation.ValidateStruct(&c, rules...)
}

func Modules(
	cfg *ModulesConfig,
	log *slog.Logger,
	tel telemetry.Telemetry,
	tls *tls.Config,
	svc service.ServerService,
	domain string,
) ([]Controller, []modules.Notifier, error) {

	controllers := make([]Controller, 0)
	notifiers := make([]modules.Notifier, 0)

	var (
		m   interface{}
		err error
	)

	// TODO: dynamic modules registration. Something like sql drivers
	// + build tags to include/exclude some modules from build
	for _, name := range cfg.Enabled {
		switch name {

		case "telegram":
			m, err = telegram.New(&cfg.Telegram, log.With("package", "telegram"), tel, svc, domain)

		case "api":
			m, err = api.New(&cfg.API, log.With("package", "api"), tls, svc)

		case "lark":
			m, err = lark.New(&cfg.Lark, log.With("package", "lark"), tel, tls, svc, domain)

		case "slack":
			m, err = slack.New(&cfg.Slack, log.With("package", "slack"), tel, svc, domain)

		}

		if err != nil {
			return nil, nil, fmt.Errorf("fail to create module %q: %w", name, err)
		}

		if c, ok := m.(Controller); ok {
			controllers = append(controllers, c)
		}

		if n, ok := m.(modules.Notifier); ok {
			notifiers = append(notifiers, n)
		}
	}

	return controllers, notifiers, nil
}
