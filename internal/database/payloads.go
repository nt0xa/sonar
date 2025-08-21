package database

import (
	"context"
	"fmt"

	"github.com/nt0xa/sonar/internal/database/models"
)

func (db *DB) PayloadsCreate(ctx context.Context, o *models.Payload) error {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsCreate")
	defer span.End()

	o.CreatedAt = now()

	query := "" +
		"INSERT INTO payloads (subdomain, user_id, name, notify_protocols, store_events, created_at) " +
		"VALUES(:subdomain, :user_id, :name, :notify_protocols, :store_events, :created_at) RETURNING id"

	if err := db.NamedQueryRowx(ctx, query, o).Scan(&o.ID); err != nil {
		return err
	}

	for _, observer := range db.obserers {
		observer.PayloadCreated(*o)
	}

	return nil
}

func (db *DB) PayloadsUpdate(ctx context.Context, o *models.Payload) error {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsUpdate")
	defer span.End()

	return db.NamedExec(ctx,
		"UPDATE payloads SET "+
			"subdomain = :subdomain, "+
			"user_id = :user_id, "+
			"name = :name, "+
			"notify_protocols = :notify_protocols, "+
			"store_events = :store_events "+
			"WHERE id = :id", o)

}

func (db *DB) PayloadGetByID(ctx context.Context, id int64) (*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadGetByID")
	defer span.End()

	var o models.Payload

	err := db.Get(ctx, &o, "SELECT * FROM payloads WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetBySubdomain(ctx context.Context, subdomain string) (*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsGetBySubdomain")
	defer span.End()

	var o models.Payload

	err := db.Get(ctx, &o, "SELECT * FROM payloads WHERE subdomain = $1", subdomain)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetByUserAndName(ctx context.Context, userID int64, name string) (*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsGetByUserAndName")
	defer span.End()

	var o models.Payload

	err := db.Get(ctx, &o, "SELECT * FROM payloads WHERE user_id = $1 and name = $2", userID, name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsFindByUserID(ctx context.Context, userID int64) ([]*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsFindByUserID")
	defer span.End()

	res := make([]*models.Payload, 0)

	err := db.Select(ctx, &res, "SELECT * FROM payloads WHERE user_id = $1 ORDER BY created_at DESC", userID)

	return res, err
}

func (db *DB) PayloadsFindByUserAndName(
	ctx context.Context,
	userID int64,
	name string,
	page uint,
	perPage uint,
) ([]*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsFindByUserAndName")
	defer span.End()

	res := make([]*models.Payload, 0)

	if page == 0 {
		page = 1 // default page
	}

	if perPage == 0 {
		perPage = 10 // default per page
	}

	err := db.Select(ctx, &res,
		"SELECT * FROM payloads "+
			"WHERE user_id = $1 and name ILIKE $2 "+
			"ORDER BY id DESC "+
			"LIMIT $3 OFFSET $4",
		userID,
		fmt.Sprintf("%%%s%%", name),
		perPage,
		perPage*(page-1),
	)

	return res, err
}

func (db *DB) PayloadsDelete(ctx context.Context, id int64) error {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsDelete")
	defer span.End()

	var o models.Payload

	if err := db.Get(ctx, &o, "DELETE FROM payloads WHERE id = $1 RETURNING *", id); err != nil {
		return err
	}

	for _, observer := range db.obserers {
		observer.PayloadDeleted(o)
	}

	return nil
}

func (db *DB) PayloadsDeleteByNamePart(ctx context.Context, userID int64, name string) ([]*models.Payload, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsDeleteByNamePart")
	defer span.End()

	res := make([]*models.Payload, 0)

	if err := db.Select(
		ctx,
		&res,
		"DELETE FROM payloads WHERE user_id = $1 AND name ILIKE $2 RETURNING *",
		userID,
		fmt.Sprintf("%%%s%%", name),
	); err != nil {
		return nil, err
	}

	for _, observer := range db.obserers {
		for _, o := range res {
			observer.PayloadDeleted(*o)
		}
	}

	return res, nil
}

func (db *DB) PayloadsGetAllSubdomains(ctx context.Context) ([]string, error) {
	ctx, span := db.tel.TraceStart(ctx, "PayloadsGetAllSubdomains")
	defer span.End()

	res := make([]string, 0)
	err := db.Select(ctx, &res, "SELECT subdomain FROM payloads")
	return res, err
}
