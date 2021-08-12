package database

import (
	"database/sql"

	"github.com/bi-zone/sonar/internal/utils/logger"
	"github.com/jmoiron/sqlx"
)

func (db *DB) NamedQueryRowx(query string, arg interface{}) *row {
	query, args, err := db.named(query, arg)
	if err != nil {
		return &row{err: err}
	}
	return db.QueryRowx(query, args...)
}

func (db *DB) NamedExec(query string, arg interface{}) error {
	query, args, err := db.named(query, arg)
	if err != nil {
		return err
	}
	return db.Exec(query, args...)
}

func (db *DB) QueryRowx(query string, args ...interface{}) *row {
	db.logQuery(query, args...)
	return &row{Row: db.DB.QueryRowx(query, args...)}
}

func (db *DB) Exec(query string, args ...interface{}) error {
	db.logQuery(query, args...)
	res, err := db.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (db *DB) Get(dest interface{}, query string, args ...interface{}) error {
	db.logQuery(query, args...)
	return db.DB.Get(dest, query, args...)
}

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	db.logQuery(query, args...)
	return db.DB.Select(dest, query, args...)
}

func (db *DB) NamedSelect(dest interface{}, query string, arg interface{}) error {
	query, args, err := db.named(query, arg)
	if err != nil {
		return err
	}
	return db.DB.Select(dest, query, args...)
}

func (db *DB) named(query string, arg interface{}) (string, []interface{}, error) {
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

func (db *DB) logQuery(query string, args ...interface{}) {
	// TODO: enable by flag
	// db.log.Printf("%s\n%+v", query, args)
}

type row struct {
	err error
	*sqlx.Row
}

func (r *row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.Row.Scan(dest...)
}

func (db *DB) Beginx() (*Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}

	return &Tx{tx, db.log}, nil
}

//
// Transactions
//

type Tx struct {
	*sqlx.Tx
	log logger.StdLogger
}

func (tx *Tx) NamedQueryRowx(query string, arg interface{}) *row {
	query, args, err := tx.named(query, arg)
	if err != nil {
		return &row{err: err}
	}
	return tx.QueryRowx(query, args...)
}

func (tx *Tx) NamedExec(query string, arg interface{}) error {
	query, args, err := tx.named(query, arg)
	if err != nil {
		return err
	}
	return tx.Exec(query, args...)
}

func (tx *Tx) QueryRowx(query string, args ...interface{}) *row {
	tx.logQuery(query, args...)
	return &row{Row: tx.Tx.QueryRowx(query, args...)}
}

func (tx *Tx) Exec(query string, args ...interface{}) error {
	tx.logQuery(query, args...)

	res, err := tx.Tx.Exec(query, args...)
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

func (tx *Tx) Get(dest interface{}, query string, args ...interface{}) error {
	tx.logQuery(query, args...)
	return tx.Tx.Get(dest, query, args...)
}

func (tx *Tx) Select(dest interface{}, query string, args ...interface{}) error {
	tx.logQuery(query, args...)
	return tx.Tx.Select(dest, query, args...)
}

func (tx *Tx) NamedSelect(dest interface{}, query string, arg interface{}) error {
	query, args, err := tx.named(query, arg)
	if err != nil {
		return err
	}
	return tx.Tx.Select(dest, query, args...)
}

func (tx *Tx) named(query string, arg interface{}) (string, []interface{}, error) {
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

func (tx *Tx) logQuery(query string, args ...interface{}) {
	// TODO: enable by flag
	// tx.log.Printf("%s\n%+v", query, args)
}
