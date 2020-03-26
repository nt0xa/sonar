package main

import (
	"crypto/tls"
	"fmt"

	"github.com/bi-zone/sonar/internal/controller"
	"github.com/bi-zone/sonar/internal/controller/api"
	"github.com/bi-zone/sonar/internal/controller/telegram"
	"github.com/bi-zone/sonar/internal/database"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sirupsen/logrus"
)

type ControllerConfig struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
	API      api.Config      `json:"api"`
}

func (c ControllerConfig) Validate() error {
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

func GetEnabledControllers(cfg *ControllerConfig, db *database.DB, log *logrus.Logger, tlsConfig *tls.Config, domain string) ([]controller.Controller, error) {
	cc := make([]controller.Controller, 0)

	var (
		c   controller.Controller
		err error
	)

	for _, name := range cfg.Enabled {
		switch name {
		case "telegram":
			c, err = telegram.New(&cfg.Telegram, db, domain)
		case "api":
			c, err = api.New(&cfg.API, db, log, tlsConfig)
		default:
			return nil, fmt.Errorf("unknown interface %v", name)
		}

		if err != nil {
			return nil, fmt.Errorf("fail to create interface %v: %w", name, err)
		}

		cc = append(cc, c)
	}

	return cc, nil
}
