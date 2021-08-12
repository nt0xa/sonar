package database

import (
	"fmt"
	"time"

	"github.com/fatih/structs"

	"github.com/bi-zone/sonar/internal/database/models"
	"github.com/bi-zone/sonar/internal/utils"
)

var usersInnerQuery = "" +
	"SELECT users.*, " +
	"COALESCE(json_object_agg(user_params.key, user_params.value) " +
	"FILTER (WHERE user_params.key IS NOT NULL), '{}') AS params " +
	"FROM users " +
	"LEFT JOIN user_params ON user_params.user_id = users.id " +
	"GROUP BY users.id"

var usersQuery = "SELECT * FROM (" + usersInnerQuery + ") AS users %s"

func (db *DB) UsersCreate(o *models.User) error {

	o.CreatedAt = time.Now()

	if o.Params.APIToken == "" {
		token, err := utils.GenerateRandomString(16)
		if err != nil {
			return err
		}
		o.Params.APIToken = token
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	query := "" +
		"INSERT INTO users (name, is_admin, created_by, created_at) " +
		"VALUES(:name, :is_admin, :created_by, :created_at) RETURNING id"

	if err := tx.NamedQueryRowx(query, o).Scan(&o.ID); err != nil {
		tx.Rollback()
		return err
	}

	for _, f := range structs.Fields(o.Params) {
		if err := tx.Exec(
			"INSERT INTO user_params (user_id, key, value) "+
				"VALUES($1, $2, $3::TEXT)", o.ID, f.Tag("json"), f.Value()); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) UsersGetByID(id int64) (*models.User, error) {
	var o models.User

	err := db.Get(&o, fmt.Sprintf(usersQuery, "WHERE users.id = $1"), id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByName(name string) (*models.User, error) {
	var o models.User

	err := db.Get(&o, fmt.Sprintf(usersQuery, "WHERE users.name = $1"), name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByParam(key models.UserParamsKey, value interface{}) (*models.User, error) {
	var o models.User

	err := db.Get(&o,
		fmt.Sprintf(usersQuery, "WHERE users.params->>$1 = $2::TEXT"), key, value)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersDelete(id int64) error {
	return db.Exec("DELETE FROM users WHERE id = $1", id)
}

func (db *DB) UsersUpdate(o *models.User) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	err = tx.NamedExec(
		"UPDATE users SET name = :name, is_admin = :is_admin, created_by = :created_by WHERE id = :id", o)

	if err != nil {
		return err
	}

	for _, f := range structs.Fields(o.Params) {
		if err := tx.Exec(
			"UPDATE user_params SET value = $1 WHERE user_id = $2 AND key = $3",
			f.Value(), o.ID, f.Tag("json")); err != nil {
			return err
		}
	}

	return tx.Commit()
}
