package httpx

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/bi-zone/sonar/pkg/netx"
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

func NotifyHandler(notify func(*Event), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := httptest.NewRecorder()
		start := time.Now()

		defer func() {
			WriteTo(w, wr)

			conn, ok := getConn(r).(*netx.LoggingConn)
			if !ok {
				return
			}

			_, secure := conn.Conn.(*tls.Conn)

			notify(&Event{
				RemoteAddr:  conn.RemoteAddr(),
				Request:     r,
				RawRequest:  conn.R.Bytes(),
				Response:    wr.Result(),
				RawResponse: conn.W.Bytes(),
				Secure:      secure,
				ReceivedAt:  start,
			})
		}()

		next.ServeHTTP(wr, r)
	})
}

func MultipartHandler(next http.Handler, maxMemory int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart otherwise it won't be read.
		if ct := r.Header.Get("Content-Type"); strings.Contains(ct, "multipart") {
			_ = r.ParseMultipartForm(maxMemory)
		}

		next.ServeHTTP(w, r)
	})
}

// WriteTo copies everything from response recorder to action response writer.
func WriteTo(w http.ResponseWriter, wr *httptest.ResponseRecorder) {
	for k, v := range wr.HeaderMap {
		w.Header()[k] = v
	}
	w.WriteHeader(wr.Code)
	wr.Body.WriteTo(w)
}
