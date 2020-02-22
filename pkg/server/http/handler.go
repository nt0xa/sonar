package http

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/bi-zone/sonar/pkg/listener"
	"github.com/bi-zone/sonar/pkg/server"
)

var ErrNoMoreData = errors.New("no more data")

const (
	maxTokenSize = 1024 * 1024 // 1mb
)

type HandlerFunc func(net.Addr, string, []byte)

type Handler struct {
	handlerFunc server.HandlerFunc
}

func (h *Handler) Handle(ctx context.Context, conn *listener.Conn) error {
	var buf bytes.Buffer
	rr := io.TeeReader(conn, &buf)

	r := bufio.NewReader(rr)
	w := bufio.NewWriter(conn)

	scanner := bufio.NewScanner(r)
	scanner.Split((&HTTPScanner{}).Scan)
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
			if err := h.sendResponse(w); err != nil {
				return err
			}

			var proto string

			if _, ok := conn.Conn.(*tls.Conn); ok {
				proto = "HTTPS"
			} else {
				proto = "HTTP"
			}

			if h.handlerFunc != nil {
				h.handlerFunc(conn.RemoteAddr(), proto, req)
			}

			return nil
		}
	}
}

func (h *Handler) sendResponse(w *bufio.Writer) error {

	rnd, err := GenerateRandomString(8)
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

func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
