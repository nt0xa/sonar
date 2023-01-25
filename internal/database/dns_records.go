package database

import (
	"time"

	"github.com/russtone/sonar/internal/database/models"
)

func (db *DB) DNSRecordsCreate(o *models.DNSRecord) error {

	o.CreatedAt = time.Now()

	query := "" +
		"INSERT INTO dns_records (payload_id, name, type, ttl, values, strategy, last_answer, last_accessed_at, created_at, index) " +
		"VALUES(:payload_id, :name, :type, :ttl, :values, :strategy, :last_answer, :last_accessed_at, :created_at," +
		" (SELECT COALESCE(MAX(index), 0) FROM dns_records dr WHERE dr.payload_id = :payload_id) + 1) " +
		"RETURNING id, index"

	return db.NamedQueryRowx(query, o).Scan(&o.ID, &o.Index)
}

func (db *DB) DNSRecordsUpdate(o *models.DNSRecord) error {

	return db.NamedExec(
		"UPDATE dns_records SET "+
			"payload_id = :payload_id, "+
			"name = :name, "+
			"type = :type, "+
			"ttl = :ttl, "+
			"values = :values, "+
			"strategy = :strategy, "+
			"last_answer = :last_answer, "+
			"last_accessed_at = :last_accessed_at "+
			"WHERE id = :id", o)
}

func (db *DB) DNSRecordsGetByID(id int64) (*models.DNSRecord, error) {
	var o models.DNSRecord

	err := db.Get(&o, "SELECT * FROM dns_records WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) DNSRecordsGetByPayloadNameAndType(payloadID int64, name string, typ string) (*models.DNSRecord, error) {
	var o models.DNSRecord

	err := db.Get(&o,
		"SELECT * FROM dns_records WHERE payload_id = $1 AND name = $2 AND type = $3",
		payloadID, name, typ)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) DNSRecordsGetByPayloadID(payloadID int64) ([]*models.DNSRecord, error) {
	res := make([]*models.DNSRecord, 0)

	err := db.Select(&res, "SELECT * FROM dns_records WHERE payload_id = $1 ORDER BY id ASC", payloadID)

	return res, err
}

func (db *DB) DNSRecordsGetCountByPayloadID(payloadID int64) (int, error) {
	var res int

	query := "SELECT COUNT(*) FROM dns_records WHERE payload_id = $1"

	err := db.Get(&res, query, payloadID)

	return res, err
}

func (db *DB) DNSRecordsGetByPayloadIDAndIndex(payloadID int64, index int64) (*models.DNSRecord, error) {
	var o models.DNSRecord

	query := "SELECT * FROM dns_records WHERE payload_id = $1 AND index = $2 ORDER BY id ASC"
	err := db.Get(&o, query, payloadID, index)

	return &o, err
}

func (db *DB) DNSRecordsDelete(id int64) error {
	return db.Exec("DELETE FROM dns_records WHERE id = $1", id)
}
