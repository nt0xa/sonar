package dns

import (
	"github.com/fatih/structs"
	"github.com/miekg/dns"
)

type Server struct {
	addr        string
	options     *options
	handlerFunc dns.HandlerFunc
}

type Meta struct {
	Qtype string
}

func New(addr string, handlerFunc dns.HandlerFunc, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	return &Server{
		options:     &options,
		addr:        addr,
		handlerFunc: handlerFunc,
	}
}

func (s *Server) SetOption(opt Option) {
	opt(s.options)
}

func (s *Server) ListenAndServe() error {
	srv := &dns.Server{
		Addr:              s.addr,
		Net:               "udp",
		Handler:           s,
		NotifyStartedFunc: s.options.notifyStartedFunc,
	}

	return srv.ListenAndServe()
}

// ServeDNS allows handler to satisfy dns.Handler interface
func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	data, _ := r.Pack()
	meta := Meta{
		Qtype: dns.Type(r.Question[0].Qtype).String(),
	}

	s.options.notifyRequestFunc(w.RemoteAddr(), data, structs.Map(meta))

	if s.handlerFunc != nil {
		s.handlerFunc(w, r)
	}
}
