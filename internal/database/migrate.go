package database

import (
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // postgres db driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(dsn string) (uint, error) {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return 0, err
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		source,
		strings.ReplaceAll(dsn, "postgres://", "pgx5://"),
	)
	if err != nil {
		return 0, fmt.Errorf("fail to create source: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, fmt.Errorf("fail to apply: %w", err)
	}

	v, _, err := m.Version()
	if err != nil {
		return 0, err
	}

	return v, nil
}
