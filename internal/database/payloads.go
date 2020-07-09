package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bi-zone/sonar/internal/models"
)

func (db *DB) PayloadsCreate(o *models.Payload) error {

	o.CreatedAt = time.Now().UTC()

	nstmt, err := db.PrepareNamed(
		"INSERT INTO payloads (subdomain, user_id, name, notify_protocols, created_at) " +
			"VALUES(:subdomain, :user_id, :name, :notify_protocols, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) PayloadsUpdate(o *models.Payload) error {

	_, err := db.NamedExec(
		"UPDATE payloads SET "+
			"subdomain = :subdomain, "+
			"user_id = :user_id, "+
			"name = :name, "+
			"notify_protocols = :notify_protocols "+
			"WHERE id = :id", o)

	return err
}

func (db *DB) PayloadGetByID(id int64) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetBySubdomain(subdomain string) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE subdomain = $1", subdomain)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetByUserAndName(userID int64, name string) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE user_id = $1 and name = $2", userID, name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsFindByUserID(userID int64) ([]*models.Payload, error) {
	res := make([]*models.Payload, 0)

	err := db.Select(&res, "SELECT * FROM payloads WHERE user_id = $1 ORDER BY created_at DESC", userID)

	return res, err
}

func (db *DB) PayloadsFindByUserAndName(userID int64, name string) ([]*models.Payload, error) {
	res := make([]*models.Payload, 0)

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
