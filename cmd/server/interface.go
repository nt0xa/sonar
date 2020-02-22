package main

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/iface"
	"github.com/bi-zone/sonar/internal/iface/telegram"
)

type InterfaceConfig struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
}

func (c InterfaceConfig) Validate() error {
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

func GetEnabledInterfaces(cfg *InterfaceConfig, db *database.DB, domain string) ([]iface.Interface, error) {
	is := make([]iface.Interface, 0)

	var (
		i   iface.Interface
		err error
	)

	for _, name := range cfg.Enabled {
		switch name {
		case "telegram":
			i, err = telegram.New(&cfg.Telegram, db, domain)
		default:
			return nil, fmt.Errorf("unknown interface %v", name)
		}

		if err != nil {
			return nil, fmt.Errorf("fail to create interface %v: %w", name, err)
		}

		is = append(is, i)
	}

	return is, nil
}
