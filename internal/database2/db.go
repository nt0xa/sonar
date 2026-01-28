package database2

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

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

