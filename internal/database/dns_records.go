package database

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/nt0xa/sonar/internal/database/models"
)

type DNSRecordsCreateParams struct {
	PayloadID      int64
	Name           string
	Type           string
	TTL            int
	Values         []string
	Strategy       string
	LastAnswer     []string
	LastAccessedAt *time.Time
}

func (db *DB) DNSRecordsCreate(ctx context.Context, p DNSRecordsCreateParams) (*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsCreate")
	defer span.End()

	o := &models.DNSRecord{
		PayloadID:      p.PayloadID,
		Name:           p.Name,
		Type:           p.Type,
		TTL:            p.TTL,
		Values:         pq.StringArray(p.Values),
		Strategy:       p.Strategy,
		LastAnswer:     pq.StringArray(p.LastAnswer),
		LastAccessedAt: p.LastAccessedAt,
		CreatedAt:      now(),
	}

	query := "" +
		"INSERT INTO dns_records (payload_id, name, type, ttl, values, strategy, last_answer, last_accessed_at, created_at, index) " +
		"VALUES(:payload_id, :name, :type, :ttl, :values, :strategy, :last_answer, :last_accessed_at, :created_at," +
		" (SELECT COALESCE(MAX(index), 0) FROM dns_records dr WHERE dr.payload_id = :payload_id) + 1) " +
		"RETURNING id, index"

	if err := db.NamedQueryRowx(ctx, query, o).Scan(&o.ID, &o.Index); err != nil {
		return nil, err
	}

	return o, nil
}

type DNSRecordsUpdateParams struct {
	ID             int64
	PayloadID      int64
	Name           string
	Type           string
	TTL            int
	Values         []string
	Strategy       string
	LastAnswer     []string
	LastAccessedAt *time.Time
}

func (db *DB) DNSRecordsUpdate(ctx context.Context, p DNSRecordsUpdateParams) (*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsUpdate")
	defer span.End()

	o := &models.DNSRecord{
		ID:             p.ID,
		PayloadID:      p.PayloadID,
		Name:           p.Name,
		Type:           p.Type,
		TTL:            p.TTL,
		Values:         pq.StringArray(p.Values),
		Strategy:       p.Strategy,
		LastAnswer:     pq.StringArray(p.LastAnswer),
		LastAccessedAt: p.LastAccessedAt,
	}

	if err := db.NamedExec(ctx,
		"UPDATE dns_records SET "+
			"payload_id = :payload_id, "+
			"name = :name, "+
			"type = :type, "+
			"ttl = :ttl, "+
			"values = :values, "+
			"strategy = :strategy, "+
			"last_answer = :last_answer, "+
			"last_accessed_at = :last_accessed_at "+
			"WHERE id = :id", o); err != nil {
		return nil, err
	}

	return o, nil
}

func (db *DB) DNSRecordsGetByID(ctx context.Context, id int64) (*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsGetByID")
	defer span.End()

	var o models.DNSRecord

	err := db.Get(ctx, &o, "SELECT * FROM dns_records WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) DNSRecordsGetByPayloadNameAndType(ctx context.Context, payloadID int64, name string, typ string) (*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsGetByPayloadNameAndType")
	defer span.End()

	var o models.DNSRecord

	err := db.Get(ctx, &o,
		"SELECT * FROM dns_records WHERE payload_id = $1 AND name = $2 AND type = $3",
		payloadID, name, typ)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) DNSRecordsGetByPayloadID(ctx context.Context, payloadID int64) ([]*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsGetByPayloadID")
	defer span.End()

	res := make([]*models.DNSRecord, 0)

	err := db.Select(ctx, &res, "SELECT * FROM dns_records WHERE payload_id = $1 ORDER BY id ASC", payloadID)

	return res, err
}

func (db *DB) DNSRecordsGetCountByPayloadID(ctx context.Context, payloadID int64) (int, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsGetCountByPayloadID")
	defer span.End()

	var res int

	query := "SELECT COUNT(*) FROM dns_records WHERE payload_id = $1"

	err := db.Get(ctx, &res, query, payloadID)

	return res, err
}

func (db *DB) DNSRecordsGetByPayloadIDAndIndex(ctx context.Context, payloadID int64, index int64) (*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsGetByPayloadIDAndIndex")
	defer span.End()

	var o models.DNSRecord

	query := "SELECT * FROM dns_records WHERE payload_id = $1 AND index = $2 ORDER BY id ASC"
	err := db.Get(ctx, &o, query, payloadID, index)

	return &o, err
}

func (db *DB) DNSRecordsDelete(ctx context.Context, id int64) error {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsDelete")
	defer span.End()

	return db.Exec(ctx, "DELETE FROM dns_records WHERE id = $1", id)
}

func (db *DB) DNSRecordsDeleteAllByPayloadID(ctx context.Context, payloadID int64) ([]*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsDeleteAllByPayloadID")
	defer span.End()

	res := make([]*models.DNSRecord, 0)

	if err := db.Select(ctx, &res,
		"DELETE FROM dns_records WHERE payload_id = $1 RETURNING *", payloadID); err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DB) DNSRecordsDeleteAllByPayloadIDAndName(ctx context.Context, payloadID int64, name string) ([]*models.DNSRecord, error) {
	ctx, span := db.tel.TraceStart(ctx, "DNSRecordsDeleteAllByPayloadIDAndName")
	defer span.End()

	res := make([]*models.DNSRecord, 0)

	if err := db.Select(ctx, &res,
		"DELETE FROM dns_records WHERE payload_id = $1 AND name = $2 RETURNING *", payloadID, name); err != nil {
		return nil, err
	}

	return res, nil
}
