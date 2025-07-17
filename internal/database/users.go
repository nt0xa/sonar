package database

import (
	"context"
	"fmt"

	"github.com/fatih/structs"

	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils"
)

var usersInnerQuery = "" +
	"SELECT users.*, " +
	"COALESCE(json_object_agg(user_params.key, user_params.value) " +
	"FILTER (WHERE user_params.key IS NOT NULL), '{}') AS params " +
	"FROM users " +
	"LEFT JOIN user_params ON user_params.user_id = users.id " +
	"GROUP BY users.id"

var usersQuery = "SELECT * FROM (" + usersInnerQuery + ") AS users %s"

func (db *DB) UsersCreate(ctx context.Context, o *models.User) error {
	ctx, span := db.tel.TraceStart(ctx, "UsersCreate")
	defer span.End()

	o.CreatedAt = now()

	if o.Params.APIToken == "" {
		token, err := utils.GenerateRandomString(16)
		if err != nil {
			return err
		}
		o.Params.APIToken = token
	}

	tx, err := db.Beginx(ctx)
	if err != nil {
		return err
	}

	query := "" +
		"INSERT INTO users (name, is_admin, created_by, created_at) " +
		"VALUES(:name, :is_admin, :created_by, :created_at) RETURNING id"

	if err := tx.NamedQueryRowx(ctx, query, o).Scan(&o.ID); err != nil {
		tx.Rollback()
		return err
	}

	for _, f := range structs.Fields(o.Params) {
		// Filter zero values here, because if for example "telegram" module is disabled we will have
		// conflict error for all users with telegram id 0.
		if !f.IsZero() {
			if err := tx.Exec(ctx,
				"INSERT INTO user_params (user_id, key, value) "+
					"VALUES($1, $2, $3::TEXT)", o.ID, f.Tag("json"), f.Value()); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (db *DB) UsersGetByID(ctx context.Context, id int64) (*models.User, error) {
	ctx, span := db.tel.TraceStart(ctx, "UsersGetByID")
	defer span.End()

	var o models.User

	err := db.Get(ctx, &o, fmt.Sprintf(usersQuery, "WHERE users.id = $1"), id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByName(ctx context.Context, name string) (*models.User, error) {
	ctx, span := db.tel.TraceStart(ctx, "UsersGetByName")
	defer span.End()

	var o models.User

	err := db.Get(ctx, &o, fmt.Sprintf(usersQuery, "WHERE users.name = $1"), name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByParam(ctx context.Context, key models.UserParamsKey, value interface{}) (*models.User, error) {
	ctx, span := db.tel.TraceStart(ctx, "UsersGetByParam")
	defer span.End()

	var o models.User

	err := db.Get(ctx, &o,
		fmt.Sprintf(usersQuery, "WHERE users.params->>$1 = $2::TEXT"), key, value)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersDelete(ctx context.Context, id int64) error {
	ctx, span := db.tel.TraceStart(ctx, "UsersDelete")
	defer span.End()

	return db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
}

func (db *DB) UsersUpdate(ctx context.Context, o *models.User) error {
	ctx, span := db.tel.TraceStart(ctx, "UsersUpdate")
	defer span.End()

	tx, err := db.Beginx(ctx)
	if err != nil {
		return err
	}

	err = tx.NamedExec(ctx,
		"UPDATE users SET name = :name, is_admin = :is_admin, created_by = :created_by WHERE id = :id", o)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, f := range structs.Fields(o.Params) {
		// Filter zero values here, because if for example "telegram" module is disabled we will have
		// conflict error for all users with telegram id 0.
		if !f.IsZero() {
			if err := tx.Exec(ctx,
				"UPDATE user_params SET value = $1 WHERE user_id = $2 AND key = $3",
				f.Value(), o.ID, f.Tag("json")); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}
