package api2

import (
	"log/slog"

	"github.com/nt0xa/sonar/internal/service"
)

type API struct {
	cfg *Config
	log *slog.Logger

	// ServerService (not plain Service) because the auth middleware needs the
	// AuthContext* identity resolvers.
	svc service.ServerService
}

func New(
	cfg *Config,
	log *slog.Logger,
	svc service.ServerService,
) (*API, error) {
	return &API{
		cfg: cfg,
		log: log,
		svc: svc,
	}, nil
}
