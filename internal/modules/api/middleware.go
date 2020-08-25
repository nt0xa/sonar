package api

import (
	"net/http"
	"strings"

	"github.com/bi-zone/sonar/internal/database/dbactions"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/errors"
)

func (api *API) checkAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")

			if token == "" {
				handleError(api.log, w, r, errors.Unauthorizedf("empty token"))
				return
			}

			u, err := api.db.UsersGetByParam(models.UserAPIToken, token)

			if err != nil {
				handleError(api.log, w, r, errors.Unauthorizedf("invalid token"))
				return
			}

			ctx := dbactions.SetUser(r.Context(), u)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (api *API) checkIsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		u, err := dbactions.GetUser(r.Context())
		if err != nil {
			handleError(api.log, w, r, err)
			return
		}

		if !u.IsAdmin {
			handleError(api.log, w, r, errors.Unauthorizedf("admin only"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
