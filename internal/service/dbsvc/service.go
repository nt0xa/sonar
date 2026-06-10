package dbsvc

import (
	"log/slog"

	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/service"
)

type Service struct {
	db  *database.DB
	log *slog.Logger
}

func New(db *database.DB, log *slog.Logger) service.Service {
	return &Service{
		db:  db,
		log: log,
	}
}
