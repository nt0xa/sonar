package server

import (
	"crypto/tls"
	"fmt"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/telegram"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sirupsen/logrus"
)

type Controller interface {
	Start() error
}

type Notifier interface {
	Notify(*models.User, *models.Payload, *models.Event) error
}

type ModulesConfig struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
	API      api.Config      `json:"api"`
}

func (c ModulesConfig) Validate() error {
	rules := make([]*validation.FieldRules, 0)
	rules = append(rules, validation.Field(&c.Enabled,
		validation.Each(validation.In("telegram", "api"))))

	for _, name := range c.Enabled {
		switch name {

		case "telegram":
			rules = append(rules, validation.Field(&c.Telegram))

		case "api":
			rules = append(rules, validation.Field(&c.API))
		}
	}

	return validation.ValidateStruct(&c, rules...)
}

func Modules(
	cfg *ModulesConfig,
	db *database.DB,
	log *logrus.Logger,
	tls *tls.Config,
	actions actions.Actions,
	domain string,
) ([]Controller, []Notifier, error) {

	controllers := make([]Controller, 0)
	notifiers := make([]Notifier, 0)

	var (
		m   interface{}
		err error
	)

	for _, name := range cfg.Enabled {
		switch name {

		case "telegram":
			m, err = telegram.New(&cfg.Telegram, db, actions, domain)

		case "api":
			m, err = api.New(&cfg.API, db, log, tls, actions)

		}

		if err != nil {
			return nil, nil, fmt.Errorf("fail to create module %q: %w", name, err)
		}

		if c, ok := m.(Controller); ok {
			controllers = append(controllers, c)
		}

		if n, ok := m.(Notifier); ok {
			notifiers = append(notifiers, n)
		}

	}

	return controllers, notifiers, nil
}
