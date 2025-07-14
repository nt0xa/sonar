package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/nt0xa/sonar/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (db *DB) NamedQueryRowx(ctx context.Context, query string, arg any) *row {
	query, args, err := db.named(query, arg)
	if err != nil {
		return &row{err: err}
	}
	return db.QueryRowx(ctx, query, args...)
}

func (db *DB) NamedExec(ctx context.Context, query string, arg any) error {
	query, args, err := db.named(query, arg)
	if err != nil {
		return err
	}
	return db.Exec(ctx, query, args...)
}

func (db *DB) QueryRowx(ctx context.Context, query string, args ...any) *row {
	ctx, span := db.tel.TraceStart(ctx, "sql.QueryRowx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return &row{Row: db.DB.QueryRowxContext(ctx, query, args...)}
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) error {
	ctx, span := db.tel.TraceStart(ctx, "sql.Exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	_, err := db.DB.ExecContext(ctx, query, args...)
	return err
}

func (db *DB) ExecResult(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, span := db.tel.TraceStart(ctx, "sql.ExecResult",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) Get(ctx context.Context, dest any, query string, args ...any) error {
	ctx, span := db.tel.TraceStart(ctx, "sql.Get",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return db.DB.GetContext(ctx, dest, query, args...)
}

func (db *DB) Select(ctx context.Context, dest any, query string, args ...any) error {
	ctx, span := db.tel.TraceStart(ctx, "sql.Select",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return db.DB.SelectContext(ctx, dest, query, args...)
}

func (db *DB) NamedSelect(ctx context.Context, dest any, query string, arg any) error {
	ctx, span := db.tel.TraceStart(ctx, "sql.NamedSelect",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()

	query, args, err := db.named(query, arg)
	if err != nil {
		return err
	}
	return db.DB.SelectContext(ctx, dest, query, args...)
}

func (db *DB) named(query string, arg any) (string, []any, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return "", nil, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, err
	}

	return db.Rebind(query), args, err
}

type row struct {
	err error
	*sqlx.Row
}

func (r *row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.Row.Scan(dest...)
}

func (db *DB) Beginx(ctx context.Context) (*Tx, error) {
	tx, err := db.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Tx{tx, db.log, db.tel}, nil
}

//
// Transactions
//

type Tx struct {
	*sqlx.Tx
	log *slog.Logger
	tel telemetry.Telemetry
}

func (tx *Tx) NamedQueryRowx(ctx context.Context, query string, arg any) *row {
	query, args, err := tx.named(query, arg)
	if err != nil {
		return &row{err: err}
	}
	return tx.QueryRowx(ctx, query, args...)
}

func (tx *Tx) NamedExec(ctx context.Context, query string, arg any) error {
	query, args, err := tx.named(query, arg)
	if err != nil {
		return err
	}
	return tx.Exec(ctx, query, args...)
}

func (tx *Tx) QueryRowx(ctx context.Context, query string, args ...any) *row {
	ctx, span := tx.tel.TraceStart(ctx, "sql.Tx.QueryRowx",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return &row{Row: tx.Tx.QueryRowxContext(ctx, query, args...)}
}

func (tx *Tx) Exec(ctx context.Context, query string, args ...any) error {
	ctx, span := tx.tel.TraceStart(ctx, "sql.Tx.Exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()

	res, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return sql.ErrNoRows
	}

	return nil
}

func (tx *Tx) Get(ctx context.Context, dest any, query string, args ...any) error {
	ctx, span := tx.tel.TraceStart(ctx, "sql.Tx.Get",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return tx.Tx.GetContext(ctx, dest, query, args...)
}

func (tx *Tx) Select(ctx context.Context, dest any, query string, args ...any) error {
	ctx, span := tx.tel.TraceStart(ctx, "sql.Tx.Select",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()
	return tx.Tx.SelectContext(ctx, dest, query, args...)
}

func (tx *Tx) NamedSelect(ctx context.Context, dest any, query string, arg any) error {
	ctx, span := tx.tel.TraceStart(ctx, "sql.Tx.NamedSelect",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("sql.query", query),
		))
	defer span.End()

	query, args, err := tx.named(query, arg)
	if err != nil {
		return err
	}
	return tx.Tx.SelectContext(ctx, dest, query, args...)
}

func (tx *Tx) named(query string, arg any) (string, []any, error) {
	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return "", nil, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, err
	}

	return tx.Rebind(query), args, err
}
