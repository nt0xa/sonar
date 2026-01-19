package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"time"

	"github.com/nt0xa/sonar/pkg/netx"
)

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

type NotifyFunc func(
	ctx context.Context,
	remoteAddr net.Addr,
	receivedAt *time.Time,
	secure bool,
	read []byte,
	written []byte,
	combined []byte,
	meta *Meta,
)

func NotifyHandler(notify NotifyFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		wr := httptest.NewRecorder()
		start := time.Now()

		next.ServeHTTP(wr, req)

		res := wr.Result()

		conn, ok := getConn(req).(*netx.LoggingConn)
		if !ok {
			return
		}

		conn.OnClose = func() {
			_, secure := conn.Conn.(*tls.Conn)

			reqBody, _ := io.ReadAll(req.Body)
			resBody, _ := io.ReadAll(res.Body)

			request := Request{
				Method:  req.Method,
				Proto:   req.Proto,
				Headers: req.Header,
				Host:    req.Host,
				URL:     req.URL.String(),
				Body:    base64.StdEncoding.EncodeToString(reqBody),
			}

			response := Response{
				Status:  res.StatusCode,
				Headers: res.Header,
				Body:    base64.StdEncoding.EncodeToString(resBody),
			}

			notify(
				req.Context(),
				conn.RemoteAddr(),
				&start,
				secure,
				conn.R.Bytes(),
				conn.W.Bytes(),
				slices.Concat(conn.R.Bytes(), conn.W.Bytes()),
				&Meta{
					Request:  request,
					Response: response,
				},
			)
		}

		for k, vv := range res.Header {
			for _, v := range vv {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(wr.Code)
		_, _ = w.Write(wr.Body.Bytes())
	})
}

// BodyReaderHandler reads body so it will appear in request log.
func BodyReaderHandler(next http.Handler, maxMemory int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)
	})
}
