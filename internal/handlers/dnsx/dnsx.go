package dnsx

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/fatih/structs"
	"github.com/miekg/dns"
)

type DNSX struct {
	db      *database.DB
	addr    string
	ip      net.IP
	origin  string
	options *options

	subdomainRegexp *regexp.Regexp
	static          map[string][]dns.RR
}

func New(addr, domain string, ip net.IP, db *database.DB, opts ...Option) (*DNSX, error) {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	re, err := regexp.Compile(fmt.Sprintf(`.*\.%s\.%s\.`,
		options.subdomainPattern, strings.ReplaceAll(domain, ".", "\\.")))

	if err != nil {
		return nil, fmt.Errorf("fail to compile subdomain regexp pattern: %w", err)
	}

	h := &DNSX{
		db:              db,
		addr:            addr,
		origin:          domain,
		ip:              ip,
		subdomainRegexp: re,
		options:         &options,
		static:          make(map[string][]dns.RR),
	}

	if err := h.addStaticRecords(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *DNSX) SetOption(opt Option) {
	opt(h.options)
}

func (h *DNSX) ListenAndServe() error {
	srv := &dns.Server{
		Addr:              h.addr,
		Net:               "udp",
		Handler:           h,
		NotifyStartedFunc: h.options.notifyStartedFunc,
	}

	return srv.ListenAndServe()
}

type Meta struct {
	Qtype string
}

// ServeDNS allows handler to satisfy dnh.Handler interface
func (h *DNSX) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	data, _ := r.Pack()
	meta := Meta{
		Qtype: qtypeStr(r.Question[0].Qtype),
	}

	if h.options.notifyRequestFunc != nil {
		h.options.notifyRequestFunc(w.RemoteAddr(), data, structs.Map(meta))
	}

	h.handleFunc(w, r)
}
