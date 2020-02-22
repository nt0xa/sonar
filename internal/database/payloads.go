package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Payload struct {
	ID        int64          `db:"id"`
	UserID    int64          `db:"user_id"`
	Subdomain string         `db:"subdomain"`
	Name      string         `db:"name"`
	Handlers  pq.StringArray `db:"handlers"`
	CreatedAt time.Time      `db:"created_at"`
}

func (db *DB) PayloadsCreate(o *Payload) error {

	o.CreatedAt = time.Now()

	nstmt, err := db.PrepareNamed(
		"INSERT INTO payloads (subdomain, user_id, name, created_at) " +
			"VALUES(:subdomain, :user_id, :name, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) PayloadsGetBySubdomain(subdomain string) (*Payload, error) {
	var o Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE subdomain = $1", subdomain)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetByUserAndName(userID int64, name string) (*Payload, error) {
	var o Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE user_id = $1 and name = $2", userID, name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsFindByUserID(userID int64) ([]*Payload, error) {
	res := make([]*Payload, 0)

	err := db.Select(&res, "SELECT * FROM payloads WHERE user_id = $1 ORDER BY created_at DESC", userID)

	return res, err
}

func (db *DB) PayloadsFindByUserAndName(userID int64, name string) ([]*Payload, error) {
	res := make([]*Payload, 0)

	err := db.Select(&res,
		"SELECT * FROM payloads WHERE user_id = $1 and name ILIKE $2 ORDER BY created_at DESC",
		userID,
		fmt.Sprintf("%%%s%%", name),
	)

	return res, err
}

func (db *DB) PayloadsDelete(id int64) error {
	res, err := db.Exec("DELETE FROM payloads WHERE id = $1", id)

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
