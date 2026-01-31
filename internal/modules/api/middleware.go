package api

import (
	"net/http"
	"strings"

	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/utils/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (api *API) checkAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")

			if token == "" {
				api.handleError(w, r, errors.Unauthorizedf("empty token"))
				return
			}

			u, err := api.db.UsersGetByAPIToken(r.Context(), token)

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

func (api *API) telemetry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := api.tel.TraceStart(r.Context(), "api",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
			),
		)
		next.ServeHTTP(w, r.WithContext(ctx))
		span.End()
	})
}
