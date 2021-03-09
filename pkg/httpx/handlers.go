package httpx

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/fatih/structs"

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

func NotifyHandler(notify func(net.Addr, []byte, map[string]interface{}), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			conn, ok := getConn(r).(*netx.LoggingConn)
			if !ok {
				return
			}

			meta := Meta{}
			_, meta.TLS = conn.Conn.(*tls.Conn)

			notify(conn.RemoteAddr(), conn.RW.Bytes(), structs.Map(meta))
		}()

		next.ServeHTTP(w, r)
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
