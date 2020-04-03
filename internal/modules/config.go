package modules

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/modules/api"
	"github.com/bi-zone/sonar/internal/modules/telegram"
)

type Config struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
	API      api.Config      `json:"api"`
}

func (c Config) Validate() error {
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
