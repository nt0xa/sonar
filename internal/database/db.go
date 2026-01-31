package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

//go:embed migrations/*.sql
var migrations embed.FS

var ErrNoRows = pgx.ErrNoRows

type DB struct {
	*Queries
	pool *pgxpool.Pool
}

func NewWithDSN(dsn string) (*DB, error) {
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool, Queries: New(pool)}, nil
}

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

func (db DB) DB() *sql.DB {
	return sql.OpenDB(stdlib.GetPoolConnector(db.pool))
}

type txFunc func(ctx context.Context, db Querier) error

func (db *DB) RunInTx(ctx context.Context, fn txFunc) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx) //nolint: errcheck // Safe to ignore

	if err := fn(ctx, db.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func RunInTx[T any](
	ctx context.Context,
	db *DB,
	fn func(ctx context.Context, db Querier) (*T, error),
) (*T, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx) //nolint: errcheck // Safe to ignore

	res, err := fn(ctx, db.WithTx(tx))
	if err != nil {
		return nil, err
	}

	return res, tx.Commit(ctx)
}

// EventsMeta contains protocol-specific event metadata.
// Only one of the protocol-specific fields will be populated per event.
type EventsMeta struct {
	// Protocol-specific metadata (only one is populated per event)
	DNS  *dnsx.Meta  `json:"dns,omitempty"`
	HTTP *httpx.Meta `json:"http,omitempty"`
	SMTP *smtpx.Meta `json:"smtp,omitempty"`
	FTP  *ftpx.Meta  `json:"ftp,omitempty"`

	// Common fields across protocols
	Secure bool `json:"secure,omitempty"`

	// GeoIP information (populated by event handler for all protocols)
	GeoIP *geoipx.Meta `json:"geoip,omitempty"`
}

type HTTPHeaders = map[string][]string
