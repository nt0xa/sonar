package database

import (
	"embed"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"

	"github.com/nt0xa/sonar/internal/utils/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	*sqlx.DB
	log      logger.StdLogger
	obserers []Observer
}

func New(cfg *Config, log logger.StdLogger) (*DB, error) {

	db, err := sqlx.Connect("postgres", cfg.DSN)

	if err != nil {
		return nil, fmt.Errorf("new: fail to connect to database: %w", err)
	}

	return &DB{
		DB:       db,
		log:      log,
		obserers: make([]Observer, 0),
	}, nil
}

func (db *DB) Migrate() error {
	fs, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: fail to create source: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: fail to create driver: %w", err)
	}

	migrations, err := migrate.NewWithInstance("iofs", fs, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: fail to init: %w", err)
	}

	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: fail to apply: %w", err)
	}

	return nil
}

func (db *DB) Observe(observer Observer) {
	db.obserers = append(db.obserers, observer)
}
