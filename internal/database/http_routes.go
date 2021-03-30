package database

import (
	"fmt"
	"time"

	"github.com/bi-zone/sonar/internal/models"
)

func (db *DB) HTTPRoutesCreate(o *models.HTTPRoute) error {

	o.CreatedAt = time.Now()

	query := "" +
		"INSERT INTO http_routes (payload_id, method, path, code, headers, body, is_dynamic, created_at) " +
		"VALUES(:payload_id, :method, :path, :code, :headers, :body, :is_dynamic, :created_at) " +
		"RETURNING id, " +
		"(SELECT COUNT(*) FROM http_routes hr WHERE hr.payload_id = :payload_id) + 1 AS index"

	nstmt, err := db.PrepareNamed(query)

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID, &o.Index)
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

	_, err := db.NamedExec(query, o)

	return err
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

	query := "SELECT *, " +
		"ROW_NUMBER() OVER (PARTITION BY payload_id ORDER BY id ASC) AS index " +
		"FROM http_routes WHERE payload_id = $1"

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

	query := "SELECT *, " +
		"ROW_NUMBER() OVER (PARTITION BY payload_id ORDER BY id ASC) AS index " +
		"FROM http_routes WHERE payload_id = $1"

	query = fmt.Sprintf("SELECT * FROM (%s) AS subq WHERE index = $2", query)

	err := db.Get(&o, query, payloadID, index)

	return &o, err
}

func (db *DB) HTTPRoutesDelete(id int64) error {
	_, err := db.Exec("DELETE FROM http_routes WHERE id = $1", id)
	return err
}
