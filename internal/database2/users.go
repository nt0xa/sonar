package database2

import (
	"context"
	"iter"
	"reflect"
)

const (
	UserTelegramID string = "telegram.id"
	UserAPIToken   string = "api.token"
	UserLarkID     string = "lark.userid"
	UserSlackID    string = "slack.id"
)

var UserParamKeys = []string{
	UserTelegramID,
	UserAPIToken,
	UserLarkID,
	UserSlackID,
}

type UserParams struct {
	TelegramID string `json:"telegram.id"`
	APIToken   string `json:"api.token"`
	LarkUserID string `json:"lark.userid"`
	SlackID    string `json:"slack.id"`
}

func (up UserParams) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		v := reflect.ValueOf(up)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			key := t.Field(i).Tag.Get("json")
			value := v.Field(i).Interface().(string)

			if !yield(key, value) {
				return
			}
		}
	}
}

type UsersCreateParams struct {
	Name      string     `db:"name"`
	IsAdmin   bool       `db:"is_admin"`
	CreatedBy *int64     `db:"created_by"`
	Params    UserParams `db:"params"`
}

func (db *DB) UsersCreate(ctx context.Context, arg UsersCreateParams) (*UsersFull, error) {
	return RunInTx(ctx, db, func(ctx context.Context, db Querier) (*UsersFull, error) {
		user, err := db.usersInsert(ctx, usersInsertParams{
			Name:      arg.Name,
			IsAdmin:   arg.IsAdmin,
			CreatedBy: arg.CreatedBy,
		})
		if err != nil {
			return nil, err
		}

		for key, value := range arg.Params.Iter() {
			if err := db.userParamsInsert(ctx, userParamsInsertParams{
				UserID: user.ID,
				Key:    key,
				Value:  value,
			}); err != nil {
				return nil, err
			}
		}

		return &UsersFull{
			ID:        user.ID,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
			CreatedBy: user.CreatedBy,
			Params:    arg.Params,
		}, nil
	})
}

type UsersUpdateParams struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	IsAdmin   bool       `db:"is_admin"`
	CreatedBy *int64     `db:"created_by"`
	Params    UserParams `db:"params"`
}

func (db *DB) UsersUpdate(ctx context.Context, arg UsersUpdateParams) (*UsersFull, error) {
	return RunInTx(ctx, db, func(ctx context.Context, db Querier) (*UsersFull, error) {
		user, err := db.usersUpdate(ctx, usersUpdateParams{
			ID:        arg.ID,
			Name:      arg.Name,
			IsAdmin:   arg.IsAdmin,
			CreatedBy: arg.CreatedBy,
		})
		if err != nil {
			return nil, err
		}

		for key, value := range arg.Params.Iter() {
			if err := db.userParamsUpdate(ctx, userParamsUpdateParams{
				UserID: user.ID,
				Key:    key,
				Value:  value,
			}); err != nil {
				return nil, err
			}
		}

		return &UsersFull{
			ID:        user.ID,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
			CreatedBy: user.CreatedBy,
			Params:    arg.Params,
		}, nil
	})
}
