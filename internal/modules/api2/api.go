package api2

import (
	"log/slog"

	"github.com/nt0xa/sonar/internal/service"
)

type API struct {
	cfg *Config
	log *slog.Logger
	svc service.Service
}

func New(
	cfg *Config,
	log *slog.Logger,
	svc service.Service,
) (*API, error) {
	return &API{
		cfg:     cfg,
		log:     log,
		svc: svc,
	}, nil
}
