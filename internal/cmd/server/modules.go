package server

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/nt0xa/sonar/internal/actions"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/modules"
	"github.com/nt0xa/sonar/internal/modules/api"
	"github.com/nt0xa/sonar/internal/modules/lark"
	"github.com/nt0xa/sonar/internal/modules/telegram"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

type Controller interface {
	Start() error
}

type ModulesConfig struct {
	Enabled []string `mapstructure:"enabled"`

	// TODO: dynamic modules config (something like json.RawMessage) to be able to not include
	// unnecessary modules in binary.
	Telegram telegram.Config `mapstructure:"telegram"`
	API      api.Config      `mapstructure:"api"`
	Lark     lark.Config     `mapstructure:"lark"`
}

func (c ModulesConfig) Validate() error {
	rules := make([]*validation.FieldRules, 0)
	rules = append(rules, validation.Field(&c.Enabled,
		validation.Each(validation.In("telegram", "api", "lark"))))

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
	db *database.DB,
	log *slog.Logger,
	tel telemetry.Telemetry,
	tls *tls.Config,
	actions actions.Actions,
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
			m, err = telegram.New(&cfg.Telegram, db, actions, domain)

		case "api":
			m, err = api.New(&cfg.API, db, log, tls, actions)

		case "lark":
			m, err = lark.New(&cfg.Lark, db, tls, actions, domain)

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
