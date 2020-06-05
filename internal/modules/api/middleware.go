package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

type contextKey string

const (
	userKey contextKey = "user"
)

func (api *API) checkAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")

			if token == "" {
				handleError(api.log, w, r, errors.Unauthorizedf("empty token"))
				return
			}

			u, err := api.db.UsersGetByParams(&database.UserParams{
				APIToken: token,
			})

			if err != nil {
				handleError(api.log, w, r, errors.Unauthorizedf("invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), userKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getUser(ctx context.Context) (*database.User, error) {
	u, ok := ctx.Value(userKey).(*database.User)
	if !ok {
		return nil, errors.Internalf("no %q key in context", userKey)
	}
	return u, nil
}
