package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/nt0xa/sonar/pkg/netx"
)

// Response represents simplified version of HTTP response.
type Response struct {
	Code    int
	Headers http.Header
	Body    io.ReadCloser
}

func MaxBytesHandler(h http.Handler, n int64) http.Handler {
	return &maxBytesHandler{h, n}
}

type maxBytesHandler struct {
	h http.Handler
	n int64
}

func (h *maxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.n)
	h.h.ServeHTTP(w, r)
}

// Event represents HTTP event.
type Event struct {
	// RemoteAddr is the address of client.
	RemoteAddr net.Addr

	// Request is a HTTP request.
	Request *http.Request

	// RawRequest is raw HTTP request.
	RawRequest []byte

	// Response is a HTTP response.
	Response *http.Response

	// RawResponse is raw HTTP response.
	RawResponse []byte

	// Secure shows if connection is TLS.
	Secure bool

	// ReceivedAt is the time of receiving query.
	ReceivedAt time.Time
}

func NotifyHandler(notify func(context.Context, *Event), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := httptest.NewRecorder()
		start := time.Now()

		next.ServeHTTP(wr, r)

		res := wr.Result()

		conn, ok := getConn(r).(*netx.LoggingConn)
		if !ok {
			return
		}

		conn.OnClose = func() {
			_, secure := conn.Conn.(*tls.Conn)

			notify(r.Context(), &Event{
				RemoteAddr:  conn.RemoteAddr(),
				Request:     r,
				RawRequest:  conn.R.Bytes(),
				Response:    res,
				RawResponse: conn.W.Bytes(),
				Secure:      secure,
				ReceivedAt:  start,
			})
		}

		for k, vv := range res.Header {
			for _, v := range vv {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(wr.Code)
		w.Write(wr.Body.Bytes())
	})
}

// BodyReaderHandler reads body so it will appear in request log.
func BodyReaderHandler(next http.Handler, maxMemory int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)
	})
}
