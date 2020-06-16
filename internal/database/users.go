package database

import (
	"database/sql"
	"time"

	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils"
)

func (db *DB) UsersCreate(o *models.User) error {

	o.CreatedAt = time.Now()

	if o.Params.APIToken == "" {
		token, err := utils.GenerateRandomString(16)
		if err != nil {
			return err
		}
		o.Params.APIToken = token
	}

	nstmt, err := db.PrepareNamed(
		"INSERT INTO users (name, params, created_at) VALUES(:name, :params, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) UsersGetByID(id int64) (*models.User, error) {
	var o models.User

	err := db.Get(&o, "SELECT * FROM users WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByName(name string) (*models.User, error) {
	var o models.User

	err := db.Get(&o, "SELECT * FROM users WHERE name = $1", name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByParams(p *models.UserParams) (*models.User, error) {
	var o models.User

	err := db.Get(&o, "SELECT * FROM users WHERE params @> $1::jsonb", p)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersDelete(id int64) error {
	res, err := db.Exec("DELETE FROM users WHERE id = $1", id)

	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return sql.ErrNoRows
	}

	return nil
}

func (db *DB) UsersUpdate(o *models.User) error {
	res, err := db.NamedExec(
		"UPDATE users SET name = :name, params = :params WHERE id = :id", o)

	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return sql.ErrNoRows
	}

	return nil
}
