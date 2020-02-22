package dns

import (
	"net"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/pkg/server"
)

type Server struct {
	*dns.Server
	*handler
	options *options
}

func NewServer(addr, domain string, ip net.IP, handlerFunc server.HandlerFunc, opts ...Option) *Server {
	options := defaultOptions

	for _, fopt := range opts {
		fopt(&options)
	}

	handler := &handler{
		domain:      domain,
		ip:          ip,
		ttl:         options.ttl,
		soa:         0,
		handlerFunc: handlerFunc,
	}

	return &Server{
		Server: &dns.Server{
			Addr:              addr,
			Net:               "udp",
			Handler:           handler,
			NotifyStartedFunc: options.notifyStartedFunc,
		},
		handler: handler,
		options: &options,
	}
}
