package api

import (
	"net/http"
)

func (api *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /profile", api.ProfileGet)

	mux.HandleFunc("GET /payloads", api.PayloadsList)
	mux.HandleFunc("POST /payloads", api.PayloadsCreate)
	mux.HandleFunc("DELETE /payloads", api.PayloadsClear)
	mux.HandleFunc("DELETE /payloads/{name}", api.PayloadsDelete)
	mux.HandleFunc("PATCH /payloads/{name}", api.PayloadsUpdate)

	mux.HandleFunc("POST /dns-records", api.DNSRecordsCreate)
	mux.HandleFunc("GET /dns-records/{payload}", api.DNSRecordsList)
	mux.HandleFunc("DELETE /dns-records/{payload}", api.DNSRecordsClear)
	mux.HandleFunc("DELETE /dns-records/{payload}/{index}", api.DNSRecordsDelete)

	mux.HandleFunc("POST /http-routes", api.HTTPRoutesCreate)
	mux.HandleFunc("GET /http-routes/{payload}", api.HTTPRoutesList)
	mux.HandleFunc("DELETE /http-routes/{payload}", api.HTTPRoutesClear)
	mux.HandleFunc("DELETE /http-routes/{payload}/{index}", api.HTTPRoutesDelete)
	mux.HandleFunc("PATCH /http-routes/{payload}/{index}", api.HTTPRoutesUpdate)

	mux.HandleFunc("GET /events/{payload}", api.EventsList)
	mux.HandleFunc("GET /events/{payload}/{index}", api.EventsGet)

	// Admin-only routes.
	admin := func(h http.HandlerFunc) http.Handler {
		return api.checkIsAdmin(h)
	}

	mux.Handle("POST /users", admin(api.UsersCreate))
	mux.Handle("DELETE /users/{name}", admin(api.UsersDelete))

	mux.Handle("GET /audit-records", admin(api.AuditRecordsList))
	mux.Handle("GET /audit-records/{id}", admin(api.AuditRecordsGet))

	return api.checkAuth(mux)
}
