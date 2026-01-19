package dnsx

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v3/challenge"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/miekg/dns"
)

type HandlerFunc func(context.Context, dns.ResponseWriter, *dns.Msg)

type Handler interface {
	ServeDNS(context.Context, dns.ResponseWriter, *dns.Msg)
}

func (f HandlerFunc) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	f(ctx, w, r)
}

// RecordGetter is an interface which must be implemented by any
// records providers like database records, in-memory records, etc.
type RecordGetter interface {
	Get(ctx context.Context, name string, qtype uint16) ([]dns.RR, error)
}

// RecordSetHandler wraps RecordGetter interface and implements
// Handlerinterface using it.
func RecordSetHandler(set RecordGetter) Handler {
	return &recordSetHandler{set}
}

type recordSetHandler struct {
	set RecordGetter
}

// ServeDNS allows recordSetHandler to implement Handlerinterface.
func (h *recordSetHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	q := r.Question[0]

	rrs, err := h.set.Get(ctx, q.Name, q.Qtype)
	if err != nil || len(rrs) == 0 {
		handleFailed(dns.RcodeServerFailure, w, r)
		return
	}

	handleSucceed(w, r, rrs)
}

// ChainHandler tries to handle query using provided DNS records set,
// if there is no answer for the query in set it calls next dns.Handler.
func ChainHandler(set RecordGetter, next Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
		q := r.Question[0]

		rrs, err := set.Get(ctx, q.Name, q.Qtype)
		if err != nil {
			handleFailed(dns.RcodeServerFailure, w, r)
			return
		}

		if rrs != nil {
			handleSucceed(w, r, rrs)
			return
		}

		next.ServeDNS(ctx, w, r)
	})
}

type NofitifyFunc func(
	ctx context.Context,
	remoteAddr net.Addr,
	receivedAt *time.Time,
	read []byte,
	written []byte,
	combined []byte,
	meta *Meta,
)

// NotifyHandler calls notify function after processing query.
func NotifyHandler(notify NofitifyFunc, next Handler) Handler {
	return HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
		wr := NewRecorder(w)

		defer func() {
			question := Question{
				Name: strings.Trim(wr.Msg.Question[0].Name, "."),
				Type: QtypeString(wr.Msg.Question[0].Qtype),
			}

			var answers []Answer
			written := ""

			if len(r.Answer) > 0 {
				for _, rr := range wr.Msg.Answer {
					answers = append(answers, Answer{
						Name: strings.Trim(rr.Header().Name, "."),
						Type: QtypeString(rr.Header().Rrtype),
						TTL:  rr.Header().Ttl,
					})
				}
				written += wr.Msg.Answer[0].String() + "\n"
			}

			notify(ctx,
				wr.RemoteAddr(),
				&wr.Start,
				[]byte(wr.Msg.Question[0].String()),
				[]byte(written),
				[]byte(wr.Msg.String()),
				&Meta{
					Question: question,
					Answer:   answers,
				},
			)
		}()

		next.ServeDNS(ctx, wr, r)
	})
}

func ChallengeHandler(next Handler) HandlerProvider {
	return &challengeHandler{
		values: make([]string, 0),
		next:   next,
	}
}

type HandlerProvider interface {
	Handler
	challenge.Provider
}

type challengeHandler struct {
	name   string
	values []string
	next   Handler
	mu     sync.Mutex
}

// Present allows Records to satisfy challenge.Provider interface
func (h *challengeHandler) Present(domain, token, keyAuth string) error {
	name, value := dns01.GetRecord(domain, keyAuth)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.name = name
	h.values = append(h.values, value)

	return nil
}

// CleanUp allows Records to satisfy challenge.Provider interface
func (h *challengeHandler) CleanUp(domain, token, keyAuth string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.name = ""
	h.values = make([]string, 0)

	return nil
}

// ServeDNS allows challengeHandler to satisfy Handlerinterface.
func (h *challengeHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) {
	name := strings.ToLower(r.Question[0].Name)

	if name == h.name {
		rrs := make([]dns.RR, 0)
		for _, value := range h.values {
			rrs = append(rrs, &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    60,
				},
				Txt: []string{value},
			})
		}
		handleSucceed(w, r, rrs)
		return
	}

	h.next.ServeDNS(ctx, w, r)
}

func handleSucceed(w dns.ResponseWriter, r *dns.Msg, answer []dns.RR) {
	m := new(dns.Msg)
	m.Authoritative = true
	m.SetReply(r)
	m.Answer = answer
	_ = w.WriteMsg(m)
}

func handleFailed(rcode int, w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(r, rcode)
	_ = w.WriteMsg(m)
}
