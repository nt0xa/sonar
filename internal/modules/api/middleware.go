package api

import (
	"net/http"
	"strings"

	"github.com/russtone/sonar/internal/actionsdb"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/internal/utils/errors"
)

func (api *API) checkAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")

			if token == "" {
				api.handleError(w, r, errors.Unauthorizedf("empty token"))
				return
			}

			u, err := api.db.UsersGetByParam(models.UserAPIToken, token)

			if err != nil {
				api.handleError(w, r, errors.Unauthorizedf("invalid token"))
				return
			}

			ctx := actionsdb.SetUser(r.Context(), u)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (api *API) checkIsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		u, _ := actionsdb.GetUser(r.Context())

		if u == nil || !u.IsAdmin {
			api.handleError(w, r, errors.Forbiddenf("admin only"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
