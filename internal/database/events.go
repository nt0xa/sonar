package database

import (
	"github.com/russtone/sonar/internal/database/models"
)

func (db *DB) EventsCreate(o *models.Event) error {

	o.CreatedAt = now()

	query := "" +
		"INSERT INTO events (payload_id, protocol, r, w, rw, meta, remote_addr, received_at, created_at, index) " +
		"VALUES(:payload_id, :protocol, :r, :w, :rw, :meta, :remote_addr, :received_at, :created_at," +
		" (SELECT COALESCE(MAX(index), 0) FROM events e WHERE e.payload_id = :payload_id) + 1) " +
		"RETURNING id, index"

	return db.NamedQueryRowx(query, o).Scan(&o.ID, &o.Index)
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
	Page
	Reverse bool
}

var defaultEventsListOptions = eventsListOptions{
	Page:    defaultPagination,
	Reverse: false,
}

type EventsListOption func(*eventsListOptions)

func EventsPagination(p Page) EventsListOption {
	return func(params *eventsListOptions) {
		if !p.IsZero() {
			params.Page = p
		}
	}
}

func EventsReverse(b bool) EventsListOption {
	return func(params *eventsListOptions) {
		params.Reverse = b
	}
}

func (db *DB) EventsListByPayloadID(payloadID int64, opts ...EventsListOption) ([]*models.Event, error) {
	options := defaultEventsListOptions

	for _, opt := range opts {
		opt(&options)
	}

	params := make(map[string]interface{})

	query := "SELECT * FROM (SELECT * FROM events WHERE payload_id = :payload_id ORDER BY id ASC) subq"
	params["payload_id"] = payloadID

	var order string

	if options.Reverse {
		order = "ASC"
	} else {
		order = "DESC"
	}

	query, params = paginate(
		query,
		params,
		"index",
		"WHERE",
		order,
		options.Page,
	)

	res := make([]*models.Event, 0)

	if err := db.NamedSelect(&res, query, params); err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DB) EventsGetByPayloadAndIndex(payloadID int64, index int64) (*models.Event, error) {
	query := "SELECT * FROM events WHERE payload_id = $1 AND index = $2"

	var res models.Event

	if err := db.Get(&res, query, payloadID, index); err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *DB) EventsDeleteOutOfLimit(payloadID int64, limit int) error {
	var minID int
	query := "SELECT COALESCE(MIN(id), 0) FROM (SELECT id FROM events WHERE payload_id = $1 ORDER BY id DESC LIMIT $2) q"
	if err := db.Get(&minID, query, payloadID, limit); err != nil {
		return err
	}
	return db.Exec("DELETE FROM events WHERE id < $1", minID)
}
