package main

import (
	"fmt"

	"github.com/bi-zone/sonar/internal/controller"
	"github.com/bi-zone/sonar/internal/controller/telegram"
	"github.com/bi-zone/sonar/internal/database"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ControllerConfig struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
}

func (c ControllerConfig) Validate() error {
	rules := make([]*validation.FieldRules, 0)
	rules = append(rules, validation.Field(&c.Enabled, validation.Each(validation.In("telegram"))))

	for _, name := range c.Enabled {
		switch name {

		case "telegram":
			rules = append(rules, validation.Field(&c.Telegram))
		}
	}

	return validation.ValidateStruct(&c, rules...)
}

func GetEnabledControllers(cfg *ControllerConfig, db *database.DB, domain string) ([]controller.Controller, error) {
	cc := make([]controller.Controller, 0)

	var (
		c   controller.Controller
		err error
	)

	for _, name := range cfg.Enabled {
		switch name {
		case "telegram":
			c, err = telegram.New(&cfg.Telegram, db, domain)
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
