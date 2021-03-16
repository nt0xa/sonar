package database

import (
	"time"

	"github.com/bi-zone/sonar/internal/models"
)

func (db *DB) EventsCreate(o *models.Event) error {

	o.CreatedAt = time.Now().UTC()

	nstmt, err := db.PrepareNamed("" +
		"INSERT INTO events (payload_id, protocol, r, w, rw, meta, remote_addr, received_at, created_at) " +
		"VALUES(:payload_id, :protocol, :r, :w, :rw, :meta, :remote_addr, :received_at, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) EventsGetByID(id int64) (*models.Event, error) {
	var o models.Event

	err := db.Get(&o, "SELECT * FROM events WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) EventsFindByPayloadID(payloadID int64) ([]*models.Event, error) {
	res := make([]*models.Event, 0)

	err := db.Select(&res, "SELECT * FROM events WHERE payload_id = $1 ORDER BY id DESC", payloadID)

	return res, err
}
