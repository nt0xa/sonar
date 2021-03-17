package database

import (
	"time"

	"github.com/bi-zone/sonar/internal/models"
)

func (db *DB) EventsCreate(o *models.Event) error {

	o.CreatedAt = time.Now()

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

type eventsListOptions struct {
	Pagination
}

var defaultEventsListOptions = eventsListOptions{
	Pagination: defaultPagination,
}

type EventsListOption func(*eventsListOptions)

func EventsPagination(p Pagination) EventsListOption {
	return func(params *eventsListOptions) {
		params.Pagination = p
	}
}

func (db *DB) EventsListByPayloadID(payloadID int64, opts ...EventsListOption) ([]*models.Event, error) {
	options := defaultEventsListOptions

	for _, opt := range opts {
		opt(&options)
	}

	params := make(map[string]interface{})

	query := "SELECT * FROM events"

	query += " WHERE payload_id = :payload_id"
	params["payload_id"] = payloadID

	if !options.Pagination.IsZero() {
		query, params = paginate(query, "id", params, options.Pagination)
	}

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	res := make([]*models.Event, 0)

	if err := stmt.Select(&res, params); err != nil {
		return nil, err
	}

	return res, nil
}
