package httpdb

import (
	"bytes"
	"database/sql"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-chi/chi/v5"

	"github.com/russtone/sonar/internal/database"
	"github.com/russtone/sonar/internal/database/models"
)

type Routes struct {
	DB     *database.DB
	Origin string
}

func (rr *Routes) Router(host string) (chi.Router, error) {

	// Get payload domain from "Host" header.
	parts := strings.Split(strings.Replace(host, "."+rr.Origin, "", 1), ".")
	domain := parts[len(parts)-1]

	// Find payload by domain.
	payload, err := rr.DB.PayloadsGetBySubdomain(domain)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	routes, err := rr.DB.HTTPRoutesGetByPayloadID(payload.ID)
	if err != nil {
		return nil, err
	}

	mux := chi.NewMux()

	for _, route := range routes {
		if route.Method != models.HTTPMethodAny {
			mux.MethodFunc(route.Method, route.Path, rr.handleFn(route))
		} else {
			mux.Handle(route.Path, rr.handleFn(route))
		}
	}

	return mux, nil
}

func Handler(rr *Routes, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux, err := rr.Router(r.Host)
		if err != nil || mux == nil {
			fallback.ServeHTTP(w, r)
			return
		}

		if !mux.Match(chi.NewRouteContext(), r.Method, r.URL.Path) {
			fallback.ServeHTTP(w, r)
			return
		}

		mux.ServeHTTP(w, r)
	})
}

func (rr *Routes) handleFn(route *models.HTTPRoute) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		data := &Data{r}

		for name, values := range route.Headers {
			for _, value := range values {
				if route.IsDynamic {
					if v, err := renderTemplate(value, data); err != nil {
						handleError(w, err)
						return
					} else {
						value = v
					}
				}
				w.Header().Add(name, value)
			}
		}

		body := route.Body
		if route.IsDynamic {
			if v, err := renderTemplate(string(body), data); err != nil {
				handleError(w, err)
				return
			} else {
				body = []byte(v)
			}
		}

		w.WriteHeader(route.Code)

		if len(route.Body) > 0 {
			w.Write(body)
		}
	}
}

type Data struct {
	r *http.Request
}

func (d *Data) Host() string {
	return d.r.Host
}

func (d *Data) Method() string {
	return d.r.Method
}

func (d *Data) Query(key string) string {
	return d.r.URL.Query().Get(key)
}

func (d *Data) Scheme() string {
	return d.r.URL.Scheme
}

func (d *Data) Path() string {
	return d.r.URL.Path
}

func (d *Data) RawQuery() string {
	return d.r.URL.RawQuery
}

func (d *Data) RequestURI() string {
	return d.r.URL.RequestURI()
}

func (d *Data) Form(key string) string {
	return d.r.FormValue(key)
}

func (d *Data) URLParam(key string) string {
	return chi.URLParam(d.r, key)
}

func (d *Data) Header(key string) string {
	return d.r.Header.Get(key)
}

func renderTemplate(t string, data interface{}) (string, error) {
	tpl, err := template.New("").Parse(t)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}
