package api

import (
	"net/http"
	"strings"

	"github.com/nt0xa/sonar/internal/service"
)

func (api *API) checkAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")
		if token == "" {
			api.handleError(w, r, httpError{Status: 401, Message: "Missing token"})
			return
		}

		ctx, err := api.svc.AuthContextByAPIToken(r.Context(), token)
		if err != nil {
			api.handleError(w, r, httpError{Status: 401, Message: "Invalid token"})
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (api *API) checkIsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c, ok := service.CallerFrom(r.Context()); !ok || !c.IsAdmin {
			api.handleError(w, r, httpError{Status: 403, Message: "Admin only"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
