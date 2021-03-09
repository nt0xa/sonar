package dnsx

import (
	"github.com/miekg/dns"
)

// Meta represents DNS event metadata.
type Meta struct {
	// DNS query type ("A", "AAAA", "TXT", etc.)
	Qtype string

	// Name is the name from query.
	Name string
}

type Server interface {
	ListenAndServe() error
}

// server contains parameters for running an DNS server.
type server struct {
	server *dns.Server
}

// New is convenient constructor for server.
func New(addr string, h dns.Handler, opts ...Option) Server {

	if h == nil {
		panic("dnsx: handler must not be nil")
	}

	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &server{
		server: &dns.Server{
			Addr:              addr,
			Net:               "udp",
			Handler:           h,
			NotifyStartedFunc: options.notifyStartedFunc,
		},
	}
}

// ListenAndServe starts the DNS server on configured address.
func (srv *server) ListenAndServe() error {
	return srv.server.ListenAndServe()
}
