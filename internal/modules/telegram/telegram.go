package telegram

import (
	"fmt"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/bi-zone/sonar/internal/database"
)

type Telegram struct {
	api *tgbotapi.BotAPI
	db  *database.DB
	cfg *Config

	domain string
}

func New(cfg *Config, db *database.DB, domain string) (*Telegram, error) {
	client := http.DefaultClient

	// Proxy
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("telegram invalid proxy url: %w", err)
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

	}

	tg, err := tgbotapi.NewBotAPIWithClient(cfg.Token, client)
	if err != nil {
		return nil, fmt.Errorf("telegram tgbotapi error: %w", err)
	}

	return &Telegram{tg, db, cfg, domain}, nil
}
