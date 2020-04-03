package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/database"
)

type contextKey string

const (
	userKey    contextKey = "user"
	payloadKey contextKey = "payload"
)

var (
	errGetUser    = errors.New("fail to get user from context")
	errGetPayload = errors.New("fail to get user from context")
)

func checkAuth(db *database.DB, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Trim(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), " ")

			if token == "" {
				handleError(log, w, r, NewError(401).SetMessage("Empty token"))
				return
			}

			u, err := db.UsersGetByParams(&database.UserParams{
				APIToken: token,
			})

			if err != nil {
				handleError(log, w, r, NewError(401).SetMessage("Invalid token"))
				return
			}

			ctx := context.WithValue(r.Context(), userKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setPayload(db *database.DB, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := r.Context().Value(userKey).(*database.User)
			if !ok {
				handleError(log, w, r, NewError(500).SetError(errGetUser))
				return
			}

			name := chi.URLParam(r, "payloadName")

			p, err := db.PayloadsGetByUserAndName(u.ID, name)
			if err != nil {
				handleError(log, w, r, NewError(404).
					SetMessage(fmt.Sprintf("Payload %q not found", name)))
				return
			}

			ctx := context.WithValue(r.Context(), payloadKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
