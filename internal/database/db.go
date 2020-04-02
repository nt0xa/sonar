package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
)

type DB struct {
	*sqlx.DB
	migrations string
}

func New(cfg *Config) (*DB, error) {

	db, err := sqlx.Connect("postgres", cfg.DSN)

	if err != nil {
		return nil, fmt.Errorf("new: fail to connect to database: %w", err)
	}

	return &DB{db, cfg.Migrations}, nil
}

func (db *DB) Migrate() error {
	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: fail to create driver: %w", err)
	}

	migrations, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", db.migrations),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: fail to create source: %w", err)
	}

	if err := migrations.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: fail to apply: %w", err)
	}

	return nil
}
