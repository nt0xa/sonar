package webx

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type Router struct {
	*http.ServeMux
	prefix string
	mws    []Middleware
}

func NewRouter(mw ...Middleware) *Router {
	return &Router{
		ServeMux: new(http.ServeMux),
		prefix:   "",
		mws:      mw,
	}
}

func (r *Router) Use(mw ...Middleware) {
	r.mws = append(r.mws, mw...)
}

func (r *Router) Group(fn func(r *Router)) {
	fn(&Router{
		ServeMux: r.ServeMux,
		prefix:   r.prefix,
		mws:      slices.Clone(r.mws),
	})
}

func (r *Router) Route(prefix string, fn func(r *Router)) {
	fn(&Router{
		ServeMux: r.ServeMux,
		prefix:   r.prefix + prefix,
		mws:      slices.Clone(r.mws),
	})
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

func (r *Router) Patch(path string, handler http.HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

func (r *Router) handle(
	method, path string,
	handler http.Handler,
) {
	fullPath := strings.TrimSuffix(r.prefix+path, "/")
	if fullPath == "" {
		fullPath = "/"
	}
	r.Handle(fmt.Sprintf("%s %s", method, fullPath), r.wrap(handler))
}

func (r *Router) wrap(handler http.Handler) http.Handler {
	out := handler
	mws := slices.Clone(r.mws)

	slices.Reverse(mws)

	for _, mw := range mws {
		out = mw(out)
	}

	return out
}
