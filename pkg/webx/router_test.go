package webx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nt0xa/sonar/pkg/webx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okHandler writes the given body with 200 OK.
func okHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}
}

// tagMiddleware appends tag to the X-Trace response header so middleware
// ordering can be observed in tests.
func tagMiddleware(tag string) webx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Trace", tag)
			next.ServeHTTP(w, r)
		})
	}
}

// do performs a request against the router and returns the recorder.
func do(t *testing.T, r *webx.Router, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestNewRouter(t *testing.T) {
	r := webx.NewRouter()
	require.NotNil(t, r)

	// A router with no middleware serves a registered route unwrapped.
	r.Get("/foo", okHandler("ok"))
	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
	assert.Empty(t, rec.Header().Values("X-Trace"))
}

func TestNewRouter_WithMiddleware(t *testing.T) {
	r := webx.NewRouter(tagMiddleware("a"))
	r.Get("/foo", okHandler("ok"))

	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, []string{"a"}, rec.Header().Values("X-Trace"))
}

func TestRouter_Methods(t *testing.T) {
	tests := []struct {
		name     string
		register func(r *webx.Router, path string, h http.HandlerFunc)
		method   string
	}{
		{"Get", (*webx.Router).Get, http.MethodGet},
		{"Post", (*webx.Router).Post, http.MethodPost},
		{"Delete", (*webx.Router).Delete, http.MethodDelete},
		{"Patch", (*webx.Router).Patch, http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := webx.NewRouter()
			tt.register(r, "/foo", okHandler("ok"))

			// Matching method succeeds.
			rec := do(t, r, tt.method, "/foo")
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "ok", rec.Body.String())

			// A different method does not match the route.
			other := http.MethodPut
			if tt.method == http.MethodPut {
				other = http.MethodGet
			}
			rec = do(t, r, other, "/foo")
			assert.NotEqual(t, http.StatusOK, rec.Code)
		})
	}
}

func TestRouter_Use(t *testing.T) {
	r := webx.NewRouter()
	r.Use(tagMiddleware("a"), tagMiddleware("b"))
	r.Get("/foo", okHandler("ok"))

	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"a", "b"}, rec.Header().Values("X-Trace"))
}

func TestRouter_MiddlewareOrder(t *testing.T) {
	// Middlewares should execute in registration order (first registered is
	// the outermost handler).
	r := webx.NewRouter(tagMiddleware("first"))
	r.Use(tagMiddleware("second"))
	r.Get("/foo", okHandler("ok"))

	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, []string{"first", "second"}, rec.Header().Values("X-Trace"))
}

func TestRouter_Route_Prefix(t *testing.T) {
	r := webx.NewRouter()
	r.Route("/api", func(r *webx.Router) {
		r.Get("/users", okHandler("users"))
	})

	rec := do(t, r, http.MethodGet, "/api/users")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "users", rec.Body.String())

	// Unprefixed path must not match.
	rec = do(t, r, http.MethodGet, "/users")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestRouter_Route_Nested(t *testing.T) {
	r := webx.NewRouter()
	r.Route("/api", func(r *webx.Router) {
		r.Route("/v1", func(r *webx.Router) {
			r.Get("/ping", okHandler("pong"))
		})
	})

	rec := do(t, r, http.MethodGet, "/api/v1/ping")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}

func TestRouter_Route_DoesNotMutateParent(t *testing.T) {
	r := webx.NewRouter()
	r.Route("/api", func(sub *webx.Router) {
		sub.Get("/inner", okHandler("inner"))
	})
	// Parent prefix is unchanged, so a sibling registers at root.
	r.Get("/outer", okHandler("outer"))

	rec := do(t, r, http.MethodGet, "/outer")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "outer", rec.Body.String())
}

func TestRouter_Group_InheritsMiddleware(t *testing.T) {
	r := webx.NewRouter(tagMiddleware("base"))
	r.Group(func(g *webx.Router) {
		g.Use(tagMiddleware("group"))
		g.Get("/foo", okHandler("ok"))
	})

	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, []string{"base", "group"}, rec.Header().Values("X-Trace"))
}

func TestRouter_Group_IsolatesMiddleware(t *testing.T) {
	r := webx.NewRouter(tagMiddleware("base"))

	r.Group(func(g *webx.Router) {
		g.Use(tagMiddleware("group"))
		g.Get("/with", okHandler("ok"))
	})

	// Middleware added inside the group must not leak to routes registered
	// outside it.
	r.Get("/without", okHandler("ok"))

	rec := do(t, r, http.MethodGet, "/with")
	assert.Equal(t, []string{"base", "group"}, rec.Header().Values("X-Trace"))

	rec = do(t, r, http.MethodGet, "/without")
	assert.Equal(t, []string{"base"}, rec.Header().Values("X-Trace"))
}

func TestRouter_TrailingSlashTrimmed(t *testing.T) {
	// A trailing slash on the registered pattern is trimmed, so "/foo/"
	// registers as "/foo".
	r := webx.NewRouter()
	r.Get("/foo/", okHandler("ok"))

	rec := do(t, r, http.MethodGet, "/foo")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestRouter_RootPath(t *testing.T) {
	// Registering "/" must register the ServeMux catch-all pattern "GET /"
	// rather than being trimmed to an empty, invalid pattern.
	r := webx.NewRouter()
	r.Get("/", okHandler("root"))

	rec := do(t, r, http.MethodGet, "/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "root", rec.Body.String())
}
