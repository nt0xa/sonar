package smtpx

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	stateHelo = iota
	stateMailFrom
	stateRcptTo
	stateData
)

var (
	addrRegexp = regexp.MustCompile(`(?i)(FROM|TO):\s*(.*?)`)
	ErrQuit    = errors.New("connection closed")
)

type Message struct {
	Helo     string   `structs:"helo"`
	Ehlo     string   `structs:"ehlo"`
	MailFrom string   `structs:"mailFrom"`
	RcptTo   []string `structs:"rcptTo"`
	Data     string   `structs:"data"`
}

type Event struct {
	RemoteAddr net.Addr
	Log        []byte
	Msg        *Message
	ReceivedAt time.Time
}

type Session struct {
	domain    string
	tlsConfig *tls.Config

	onClose func(*Event)

	lines   chan string
	rdy     chan struct{}
	conn    net.Conn
	r       *bufio.Reader
	w       *bufio.Writer
	scanner *bufio.Scanner
	conv    *bytes.Buffer

	state int

	msg *Message

	mu sync.RWMutex
}

func NewSession(conn net.Conn, domain string, tlsConfig *tls.Config, onClose func(*Event)) *Session {
	var buf bytes.Buffer

	r := bufio.NewReader(io.TeeReader(conn, &buf))
	w := bufio.NewWriter(io.MultiWriter(conn, &buf))
	scanner := bufio.NewScanner(r)

	return &Session{
		domain:    domain,
		tlsConfig: tlsConfig,
		onClose:   onClose,

		lines:   make(chan string, 1),
		rdy:     make(chan struct{}, 1),
		conn:    conn,
		r:       r,
		w:       w,
		scanner: scanner,
		conv:    &buf,

		state: stateHelo,
		msg: &Message{
			RcptTo: make([]string, 0),
		},
	}
}

func (s *Session) start(ctx context.Context) error {
	start := time.Now()

	if s.onClose != nil {
		defer func() {
			s.onClose(&Event{
				RemoteAddr: s.conn.RemoteAddr(),
				Log:        s.conv.Bytes(),
				Msg:        s.msg,
				ReceivedAt: start,
			})
		}()
	}

	if err := s.greet(); err != nil {
		return err
	}

	defer close(s.lines)

	s.ready()

	for {
		select {

		case <-ctx.Done():
			return nil

		case <-s.rdy:
			s.mu.RLock()
			if !s.scanner.Scan() {
				s.mu.RUnlock()
				return s.scanner.Err()
			}
			s.mu.RUnlock()

			s.lines <- s.scanner.Text()

		case line := <-s.lines:
			if err := s.handle(line); err == ErrQuit {
				return nil
			} else if err != nil {
				return err
			}
		}
	}
}

func (s *Session) handle(line string) error {
	defer s.ready()

	cmd, args := s.parseCmd(line)

	switch s.state {

	// TODO: RSET, VRFY
	case stateHelo:
		switch cmd {
		case "HELO":
			return s.handleHelo(args)
		case "EHLO":
			return s.handleEhlo(args)
		case "QUIT":
			return s.handleQuit(args)
		case "NOOP":
			return s.handleNoop(args)
		default:
			return s.badSequenceError()
		}

	case stateMailFrom:
		switch cmd {
		case "STARTTLS":
			return s.handleStartTLS(args)
		case "MAIL":
			return s.handleMailFrom(args)
		case "QUIT":
			return s.handleQuit(args)
		case "NOOP":
			return s.handleNoop(args)
		default:
			return s.badSequenceError()
		}

	case stateRcptTo:
		switch cmd {
		case "RCPT":
			return s.handleRcptTo(args)
		case "DATA":
			return s.handleData(args)
		case "QUIT":
			return s.handleQuit(args)
		case "NOOP":
			return s.handleNoop(args)
		default:
			return s.badSequenceError()
		}

	case stateData:
		return s.handleData(line)
	}

	return nil
}

func (s *Session) parseCmd(line string) (string, string) {
	parts := strings.SplitN(line, " ", 2)

	cmd := strings.ToUpper(parts[0])
	args := ""

	if len(parts) > 1 {
		args = parts[1]
	}

	return cmd, args
}

func (s *Session) writeLine(line string) error {
	if _, err := s.w.WriteString(line + "\r\n"); err != nil {
		return err
	}
	return s.w.Flush()
}

func (s *Session) badSequenceError() error {
	return s.writeLine("503 Bad sequence of commands")
}

func (s *Session) greet() error {
	return s.writeLine(fmt.Sprintf("220 %s SMTP Server ready", s.domain))
}

func (s *Session) ready() {
	s.rdy <- struct{}{}
}

func (s *Session) handleHelo(args string) error {
	s.msg.Helo = args
	s.state = stateMailFrom
	return s.writeLine("250 Hello")
}

func (s *Session) handleNoop(args string) error {
	return s.writeLine("250 OK")
}

func (s *Session) handleEhlo(args string) error {
	s.msg.Ehlo = args
	s.state = stateMailFrom

	if s.tlsConfig == nil {
		return s.writeLine(fmt.Sprintf("250 %s", s.domain))
	}

	if err := s.writeLine(fmt.Sprintf("250-%s", s.domain)); err != nil {
		return err
	}

	return s.writeLine("250 STARTTLS")
}

func (s *Session) handleStartTLS(args string) error {
	if err := s.writeLine("220 Ready to start TLS"); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	conn := tls.Server(s.conn, s.tlsConfig)

	if err := conn.Handshake(); err != nil {
		return err
	}

	s.conn = net.Conn(conn)

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, s.conv); err != nil {
		return err
	}

	s.r = bufio.NewReader(io.TeeReader(conn, &buf))
	s.w = bufio.NewWriter(io.MultiWriter(conn, &buf))
	s.scanner = bufio.NewScanner(s.r)
	s.conv = &buf
	s.state = stateHelo

	return nil
}

func (s *Session) handleMailFrom(args string) error {
	s.msg.MailFrom = addrRegexp.ReplaceAllString(args, "$2")
	s.state = stateRcptTo
	return s.writeLine("250 OK")
}

func (s *Session) handleRcptTo(args string) error {
	s.msg.RcptTo = append(s.msg.RcptTo, addrRegexp.ReplaceAllString(args, "$2"))
	s.state = stateRcptTo
	return s.writeLine("250 OK")
}

func (s *Session) handleData(args string) error {
	if args == "" && s.state != stateData {
		s.state = stateData
		return s.writeLine("354 Send data")
	} else if args != "." {
		s.msg.Data += args + "\n"
		return nil
	}

	s.state = stateMailFrom
	return s.writeLine("250 OK")
}

func (s *Session) handleQuit(args string) error {
	_ = s.writeLine("221 OK")
	return ErrQuit
}
