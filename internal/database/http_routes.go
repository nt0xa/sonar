package database

import (
	"time"

	"github.com/bi-zone/sonar/internal/database/models"
)

func (db *DB) HTTPRoutesCreate(o *models.HTTPRoute) error {

	o.CreatedAt = time.Now()

	query := "" +
		"INSERT INTO http_routes (payload_id, method, path, code, headers, body, is_dynamic, created_at, index) " +
		"VALUES(:payload_id, :method, :path, :code, :headers, :body, :is_dynamic, :created_at, " +
		"(SELECT COALESCE(MAX(index), 0) FROM http_routes hr WHERE hr.payload_id = :payload_id) + 1) " +
		"RETURNING id, index"

	return db.NamedQueryRowx(query, o).Scan(&o.ID, &o.Index)
}

func (db *DB) HTTPRoutesUpdate(o *models.HTTPRoute) error {

	query := "" +
		"UPDATE http_routes SET " +
		"payload_id = :payload_id, " +
		"method = :method, " +
		"path = :path, " +
		"code = :code, " +
		"headers = :headers, " +
		"body = :body, " +
		"is_dynamic = :is_dynamic " +
		"WHERE id = :id"

	return db.NamedExec(query, o)
}

func (db *DB) HTTPRoutesGetByID(id int64) (*models.HTTPRoute, error) {
	var o models.HTTPRoute

	err := db.Get(&o, "SELECT * FROM http_routes WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) HTTPRoutesGetByPayloadID(payloadID int64) ([]*models.HTTPRoute, error) {
	res := make([]*models.HTTPRoute, 0)

	query := "SELECT * FROM http_routes WHERE payload_id = $1"

	err := db.Select(&res, query, payloadID)

	return res, err
}

func (db *DB) HTTPRoutesGetByPayloadMethodAndPath(payloadID int64, method string, path string) (*models.HTTPRoute, error) {
	var o models.HTTPRoute

	err := db.Get(&o,
		"SELECT * FROM http_routes WHERE payload_id = $1 AND method = $2 AND path = $3",
		payloadID, method, path)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) HTTPRoutesGetByPayloadIDAndIndex(payloadID int64, index int64) (*models.HTTPRoute, error) {
	var o models.HTTPRoute

	query := "SELECT * FROM http_routes WHERE payload_id = $1 AND index = $2"
	err := db.Get(&o, query, payloadID, index)

	return &o, err
}

func (db *DB) HTTPRoutesDelete(id int64) error {
	return db.Exec("DELETE FROM http_routes WHERE id = $1", id)
}
