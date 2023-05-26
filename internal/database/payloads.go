package database

import (
	"fmt"

	"github.com/russtone/sonar/internal/database/models"
)

func (db *DB) PayloadsCreate(o *models.Payload) error {

	o.CreatedAt = now()

	query := "" +
		"INSERT INTO payloads (subdomain, user_id, name, notify_protocols, store_events, created_at) " +
		"VALUES(:subdomain, :user_id, :name, :notify_protocols, :store_events, :created_at) RETURNING id"

	if err := db.NamedQueryRowx(query, o).Scan(&o.ID); err != nil {
		return err
	}

	for _, observer := range db.obserers {
		observer.PayloadCreated(*o)
	}

	return nil
}

func (db *DB) PayloadsUpdate(o *models.Payload) error {

	return db.NamedExec(
		"UPDATE payloads SET "+
			"subdomain = :subdomain, "+
			"user_id = :user_id, "+
			"name = :name, "+
			"notify_protocols = :notify_protocols, "+
			"store_events = :store_events "+
			"WHERE id = :id", o)

}

func (db *DB) PayloadGetByID(id int64) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetBySubdomain(subdomain string) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE subdomain = $1", subdomain)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsGetByUserAndName(userID int64, name string) (*models.Payload, error) {
	var o models.Payload

	err := db.Get(&o, "SELECT * FROM payloads WHERE user_id = $1 and name = $2", userID, name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) PayloadsFindByUserID(userID int64) ([]*models.Payload, error) {
	res := make([]*models.Payload, 0)

	err := db.Select(&res, "SELECT * FROM payloads WHERE user_id = $1 ORDER BY created_at DESC", userID)

	return res, err
}

func (db *DB) PayloadsFindByUserAndName(userID int64, name string) ([]*models.Payload, error) {
	res := make([]*models.Payload, 0)

	err := db.Select(&res,
		"SELECT * FROM payloads WHERE user_id = $1 and name ILIKE $2 ORDER BY created_at DESC",
		userID,
		fmt.Sprintf("%%%s%%", name),
	)

	return res, err
}

func (db *DB) PayloadsDelete(id int64) error {
	var o models.Payload

	if err := db.Get(&o, "DELETE FROM payloads WHERE id = $1 RETURNING *", id); err != nil {
		return err
	}

	for _, observer := range db.obserers {
		observer.PayloadDeleted(o)
	}

	return nil
}

func (db *DB) PayloadsGetAllSubdomains() ([]string, error) {
	res := make([]string, 0)
	err := db.Select(&res, "SELECT subdomain FROM payloads")
	return res, err
}
