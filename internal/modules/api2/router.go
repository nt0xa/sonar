package api2

import (
	"net/http"

	"github.com/nt0xa/sonar/pkg/webx"
)

func (api *API) Handler() http.Handler {
	r := webx.NewRouter()

	r.Get("/profile", api.ProfileGet)

	r.Route("/payloads", func(r *webx.Router) {
		r.Get("/", api.PayloadsList)
		r.Post("/", api.PayloadsCreate)
		r.Delete("/", api.PayloadsClear)
		r.Route("/{name}", func(r *webx.Router) {
			r.Delete("/", api.PayloadsDelete)
			r.Patch("/", api.PayloadsUpdate)
		})
	})

	r.Route("/dns-records", func(r *webx.Router) {
		r.Post("/", api.DNSRecordsCreate)
		r.Route("/{payload}", func(r *webx.Router) {
			r.Get("/", api.DNSRecordsList)
			r.Delete("/", api.DNSRecordsClear)
			r.Route("/{index}", func(r *webx.Router) {
				r.Delete("/", api.DNSRecordsDelete)
			})
		})
	})

	r.Route("/http-routes", func(r *webx.Router) {
		r.Post("/", api.HTTPRoutesCreate)
		r.Route("/{payload}", func(r *webx.Router) {
			r.Get("/", api.HTTPRoutesList)
			r.Delete("/", api.HTTPRoutesClear)
			r.Route("/{index}", func(r *webx.Router) {
				r.Delete("/", api.HTTPRoutesDelete)
				r.Patch("/", api.HTTPRoutesUpdate)
			})
		})
	})

	r.Route("/events", func(r *webx.Router) {
		r.Route("/{payload}", func(r *webx.Router) {
			r.Get("/", api.EventsList)
			r.Get("/{index}", api.EventsGet)
		})
	})

	// r.Group(func(r *webx.Router) {
	// 	r.Use(api.checkIsAdmin)
	//
	// 	r.Route("/users", func(r *webx.Router) {
	// 		r.Post("/", api.UsersCreate)
	// 		r.Route("/{name}", func(r *webx.Router) {
	// 			r.Delete("/", api.UsersDelete)
	// 		})
	// 	})
	//
	// 	r.Route("/audit-records", func(r *webx.Router) {
	// 		r.Get("/", api.AuditRecordsList)
	// 		r.Route("/{id}", func(r *webx.Router) {
	// 			r.Get("/", api.AuditRecordsGet)
	// 		})
	// 	})
	// })

	return r
}
