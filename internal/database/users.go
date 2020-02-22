package database

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	Params    UserParams `db:"params"`
	CreatedAt time.Time  `db:"created_at"`
}

type UserParams struct {
	TelegramID int64 `json:"telegramId"`
}

func (p UserParams) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *UserParams) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

func (db *DB) UsersCreate(o *User) error {

	o.CreatedAt = time.Now()

	nstmt, err := db.PrepareNamed(
		"INSERT INTO users (name, params, created_at) VALUES(:name, :params, :created_at) RETURNING id")

	if err != nil {
		return err
	}

	return nstmt.QueryRowx(o).Scan(&o.ID)
}

func (db *DB) UsersGetByID(id int64) (*User, error) {
	var o User

	err := db.Get(&o, "SELECT * FROM users WHERE id = $1", id)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByName(name string) (*User, error) {
	var o User

	err := db.Get(&o, "SELECT * FROM users WHERE name = $1", name)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (db *DB) UsersGetByParams(p *UserParams) (*User, error) {
	var o User

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

func (db *DB) UsersUpdate(o *User) error {
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
