package database

import (
	"time"

	"github.com/bi-zone/sonar/internal/models"
)

func (db *DB) DNSRecordsCreate(o *models.DNSRecord) error {

	o.CreatedAt = time.Now()

	nstmt, err := db.PrepareNamed(
		"INSERT INTO dns_records (payload_id, name, type, ttl, values, created_at) " +
			"VALUES(:payload_id, :name, :type, :ttl, :values, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) DNSRecordsUpdate(o *models.DNSRecord) error {

	_, err := db.NamedExec(
		"UPDATE dns_records SET "+
			"payload_id = :payload_id, "+
			"name = :name, "+
			"type = :type, "+
			"ttl = :ttl, "+
			"values = :values "+
			"WHERE id = :id", o)

	return err
}

func (db *DB) DNSRecordsGetByID(id int64) (*models.DNSRecord, error) {
	var o models.DNSRecord

	err := db.Get(&o, "SELECT * FROM dns_records WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) DNSRecordsGetByPayloadNameType(payloadID int64, name string, typ string) (*models.DNSRecord, error) {
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

	err := db.Select(&res,
		"SELECT * FROM dns_records WHERE payload_id = $1",
		payloadID)

	return res, err
}

func (db *DB) DNSRecordsDelete(id int64) error {
	_, err := db.Exec("DELETE FROM dns_records WHERE id = $1", id)
	return err
}
