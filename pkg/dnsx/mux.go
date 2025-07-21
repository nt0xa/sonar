// Based on github.com/miekg/dns/serve_mux.go
package dnsx

import (
	"context"
	"sync"

	"github.com/miekg/dns"
)

type ServeMux struct {
	z map[string]Handler
	m sync.RWMutex
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func (mux *ServeMux) match(q string, t uint16) Handler {
	mux.m.RLock()
	defer mux.m.RUnlock()
	if mux.z == nil {
		return nil
	}

	q = dns.CanonicalName(q)

	var handler Handler
	for off, end := 0, false; !end; off, end = dns.NextLabel(q, off) {
		if h, ok := mux.z[q[off:]]; ok {
			if t != dns.TypeDS {
				return h
			}
			// Continue for DS to see if we have a parent too, if so delegate to the parent
			handler = h
		}
	}

	// Wildcard match, if we have found nothing try the root zone as a last resort.
	if h, ok := mux.z["."]; ok {
		return h
	}

	return handler
}

// Handle adds a handler to the ServeMux for pattern.
func (mux *ServeMux) Handle(pattern string, handler Handler) {
	if pattern == "" {
		panic("dns: invalid pattern " + pattern)
	}
	mux.m.Lock()
	if mux.z == nil {
		mux.z = make(map[string]Handler)
	}
	mux.z[dns.CanonicalName(pattern)] = handler
	mux.m.Unlock()
}

// HandleFunc adds a handler function to the ServeMux for pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(context.Context, dns.ResponseWriter, *dns.Msg)) {
	mux.Handle(pattern, HandlerFunc(handler))
}

func (mux *ServeMux) ServeDNS(ctx context.Context, w dns.ResponseWriter, req *dns.Msg) {
	var h Handler
	if len(req.Question) >= 1 { // allow more than one question
		h = mux.match(req.Question[0].Name, req.Question[0].Qtype)
	}

	if h != nil {
		h.ServeDNS(ctx, w, req)
	} else {
		handleFailed(dns.RcodeRefused, w, req)
	}
}
