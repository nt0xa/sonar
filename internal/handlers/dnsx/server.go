package dnsx

import (
	"strings"

	"github.com/fatih/structs"
	"github.com/miekg/dns"
)

// Server contains parameters for running an DNS server.
type Server struct {
	// addr is address to listen on, ":53" if empty.
	addr string

	// origin is the root domain of the server.
	origin string

	// finders is the list of Finder in descending order of priority.
	finders []Finder

	// options is the additional parameters of DNS server.
	options *options
}

// Finder abstracts different implementations of DNS answer find (static records, database records, etc.).
type Finder interface {

	// Find returns answer records for query.
	// If there are no records found should return nil.
	// Returning empty records array will stop processing.
	Find(name string, qtype uint16) ([]dns.RR, error)
}

// New returns new DNS server instance.
func New(addr, origin string, finders []Finder, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		addr:    addr,
		origin:  origin,
		finders: finders,
		options: &options,
	}
}

// SetOptions sets DNS server option.
func (srv *Server) SetOption(opt Option) {
	opt(srv.options)
}

// ListenAndServe starts the DNS server on configured address.
func (srv *Server) ListenAndServe() error {
	s := dns.Server{
		Addr:              srv.addr,
		Net:               "udp",
		Handler:           srv,
		NotifyStartedFunc: srv.options.notifyStartedFunc,
	}
	return s.ListenAndServe()
}

// Meta represents DNS event metadata.
type Meta struct {
	// DNS query type ("A", "AAAA", "TXT", etc.)
	Qtype string
}

// ServeDNS allows handler to satisfy dnh.Handler interface
func (srv *Server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	res := new(dns.Msg)

	if srv.options.notifyRequestFunc != nil {
		meta := Meta{
			Qtype: qtypeStr(req.Question[0].Qtype),
		}

		// Wrap with func() to get result response.
		defer func() {
			srv.options.notifyRequestFunc(w.RemoteAddr(), []byte(res.String()), structs.Map(meta))
		}()
	}

	res.SetReply(req)
	res.Authoritative = true

	if len(req.Question) == 0 {
		res.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(res)
		return
	}

	qname := req.Question[0].Name
	qtype := req.Question[0].Qtype

	if !strings.HasSuffix(qname, srv.origin+".") {
		res.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(res)
		return
	}

	// Construct all possible wildcards.
	// test1.test2 -> [test1.test2, *.test2, *]
	qnames := []string{qname}
	qnames = append(qnames, makeWildcards(qname)...)

loop:
	for _, finder := range srv.finders {
		for _, name := range qnames {

			rrs, err := finder.Find(name, qtype)

			if err != nil {
				res.Rcode = dns.RcodeServerFailure
				_ = w.WriteMsg(res)
				return
			}

			// Stop processing even if len(rrs) == 0.
			// This is required to be able to have only limited
			// DNS records for subdomain (e.g. only "AAAA" record without "A").
			if rrs != nil {
				res.Answer = rrs
				break loop
			}
		}
	}


	if len(res.Answer) == 0 {
		res.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(res)
		return
	}

	res.Answer = fixWildcards(res.Answer, qname)

	_ = w.WriteMsg(res)
}
