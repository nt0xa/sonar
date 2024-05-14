package database

import (
	"github.com/russtone/sonar/internal/database/models"
)

func (db *DB) EventsCreate(o *models.Event) error {

	o.CreatedAt = now()

	query := "" +
		"INSERT INTO events (payload_id, protocol, r, w, rw, meta, remote_addr, received_at, created_at) " +
		"VALUES(:payload_id, :protocol, :r, :w, :rw, :meta, :remote_addr, :received_at, :created_at)" +
		"RETURNING id"

	return db.NamedQueryRowx(query, o).Scan(&o.ID)
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

	query := "SELECT * FROM (SELECT *, ROW_NUMBER() OVER(ORDER BY id ASC) AS index FROM events WHERE payload_id = :payload_id) subq"
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
	query := "SELECT * FROM (SELECT *, ROW_NUMBER() OVER(ORDER BY id ASC) AS index FROM events WHERE payload_id = $1) subq WHERE index = $2"

	var res models.Event

	if err := db.Get(&res, query, payloadID, index); err != nil {
		return nil, err
	}

	return &res, nil
}
