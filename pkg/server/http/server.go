package http

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/pkg/listener"
)

const (
	maxTokenSize = 1024 * 1024 // 1mb
)

type Server struct {
	addr    string
	options *options
}

func New(addr string, opts ...Option) *Server {
	options := defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		addr:    addr,
		options: &options,
	}
}

func (s *Server) SetOption(opt Option) {
	opt(s.options)
}

func (s *Server) ListenAndServe() error {
	l := &listener.Listener{
		Addr:              s.addr,
		Handler:           s,
		IdleTimeout:       s.options.idleTimeout,
		SessionTimeout:    s.options.sessionTimeout,
		NotifyStartedFunc: s.options.notifyStartedFunc,
		TLSConfig:         s.options.tlsConfig,
	}

	return l.Listen()
}

func (s *Server) Handle(ctx context.Context, conn *listener.Conn) error {
	var buf bytes.Buffer
	rr := io.TeeReader(conn, &buf)

	r := bufio.NewReader(rr)
	w := bufio.NewWriter(conn)

	scanner := bufio.NewScanner(r)
	scanner.Split((&Scanner{}).Scan)
	scanner.Buffer(make([]byte, maxTokenSize), maxTokenSize)

	ch := make(chan []byte)

	go func() {
		defer close(ch)

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				log.Printf("%v (%v)", err, conn.RemoteAddr())
			}
			return
		}

		ch <- scanner.Bytes()
	}()

	for {
		select {

		case <-ctx.Done():
			// TODO: check Host header maybe?
			log.Printf("session timeout exceed for %v", conn.RemoteAddr())
			return nil

		case req := <-ch:
			if err := s.sendResponse(w); err != nil {
				return err
			}

			meta := make(map[string]interface{})

			_, meta["tls"] = conn.Conn.(*tls.Conn)

			if s.options.notifyRequestFunc != nil {
				s.options.notifyRequestFunc(conn.RemoteAddr(), req, meta)
			}

			return nil
		}
	}
}

func (s *Server) sendResponse(w *bufio.Writer) error {

	rnd, err := utils.GenerateRandomString(8)
	if err != nil {
		return err
	}

	body := fmt.Sprintf("<html><body>%s</body></html>", rnd)

	_, err = w.WriteString("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"\r\n" +
		body)

	if err != nil {
		return err
	}

	return w.Flush()
}
