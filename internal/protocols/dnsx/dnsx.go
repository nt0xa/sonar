package dnsx

import (
	"net"
	"strings"

	"github.com/fatih/structs"
	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/protocols/dnsx/dnsutils"
)

// Handler abstracts different implementations of DNS answer find (static records, database records, etc.).
type Handler interface {

	// ServeDNS returns answer for DNS query.
	ServeDNS(name string, qtype uint16) ([]dns.RR, error)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as DNS handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(string, uint16) ([]dns.RR, error)

// ServeDNS calls f(w, r).
func (f HandlerFunc) ServeDNS(name string, qtype uint16) {
	f(name, qtype)
}

// Meta represents DNS event metadata.
type Meta struct {
	// DNS query type ("A", "AAAA", "TXT", etc.)
	Qtype string

	Name string
}

// Server contains parameters for running an DNS server.
type Server struct {
	// Addr is address to listen on, ":53" if empty.
	Addr string

	// Origin is the root domain of the server.
	Origin string

	// Handler is the list of handlers in descending order of priority.
	Handlers []Handler

	// If NotifyStartedFunc is set it is called once the server has started listening.
	NotifyStartedFunc func()

	// If NotifyRequestFunc is set it is called on every DNS query.
	NotifyRequestFunc func(net.Addr, []byte, map[string]interface{})

	server *dns.Server
}

// ListenAndServe starts the DNS server on configured address.
func (srv *Server) ListenAndServe() error {
	srv.server = &dns.Server{
		Addr:              srv.Addr,
		Net:               "udp",
		Handler:           srv,
		NotifyStartedFunc: srv.NotifyStartedFunc,
	}

	return srv.server.ListenAndServe()
}

// ServeDNS allows Server to implement dns.Handler interface
func (srv *Server) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	res := new(dns.Msg)

	res.SetReply(req)
	res.Authoritative = true

	if len(req.Question) == 0 ||
		!strings.HasSuffix(strings.ToLower(req.Question[0].Name), srv.Origin+".") ||
		srv.Handlers == nil {
		res.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(res)
		return
	}

	if srv.NotifyRequestFunc != nil {
		meta := Meta{
			Qtype: dnsutils.QtypeString(req.Question[0].Qtype),
			Name:  strings.Trim(req.Question[0].Name, "."),
		}

		// Wrap with func() to capture res after modifications.
		defer func() {
			srv.NotifyRequestFunc(w.RemoteAddr(), []byte(res.String()), structs.Map(meta))
		}()
	}

	qname := req.Question[0].Name
	qtype := req.Question[0].Qtype

	// Construct all possible wildcards.
	// test1.test2 -> [test1.test2, *.test2, *]
	qname = strings.ToLower(qname)
	qnames := []string{qname}
	qnames = append(qnames, dnsutils.MakeWildcards(qname)...)

loop:
	for _, handler := range srv.Handlers {
		for _, name := range qnames {
			rrs, err := handler.ServeDNS(name, qtype)

			if err != nil {
				res.Rcode = dns.RcodeServerFailure
				_ = w.WriteMsg(res)
				return
			}

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

	// Remove wildcards in answers.
	rrs := make([]dns.RR, 0)
	for _, rr := range res.Answer {
		// In case of wildcard it is required to change name
		// as in was in the question.
		if strings.HasPrefix(rr.Header().Name, "*") {
			rr = dns.Copy(rr)
			rr.Header().Name = qname
		}

		rrs = append(rrs, rr)
	}

	res.Answer = rrs

	_ = w.WriteMsg(res)
}

// Shutdown shuts down a server. After a call to Shutdown, ListenAndServe will return.
func (srv *Server) Shutdown() error {
	return srv.server.Shutdown()
}
