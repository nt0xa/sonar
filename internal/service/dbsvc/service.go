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

var _ service.ServerService = (*Service)(nil)

func New(db *database.DB, log *slog.Logger) service.ServerService {
	return &Service{
		db:  db,
		log: log,
	}
}
