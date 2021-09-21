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
	"time"
)

// SMTP session states.
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

// Data stores args passed to the corresponding SMTP commands.
type Data struct {
	// Helo is a "HELO" command data.
	Helo string

	// Ehlo is a "EHLO" command data.
	Ehlo string

	// MailFrom is a "MAIL FROM" command data.
	MailFrom string

	// RcptTo is a "RCPT TO" command data.
	RcptTo []string

	// Data is a "DATA" command data.
	Data string
}

// Event represents SMTP event.
type Event struct {
	// RemoteAddre is remote IP address.
	RemoteAddr net.Addr

	// Log is a full session log.
	Log []byte

	// Data stores args passed to the corresponding SMTP commands during
	// a session.
	Data *Data

	// Secure shows connection was secure (with TLS).
	Secure bool

	// ReceivedAt is a session start time.
	ReceivedAt time.Time
}

// Msg is used to store provided command responses.
type Msg struct {
	// Greet is server greet message.
	Greet string

	// Ehlo is a first line of EHLO command response.
	Ehlo string
}

// session contains all data required to handle SMTP session.
type session struct {
	// messages stores provided command responses.
	messages Msg

	// tlsConfig is an optional TLS config.
	// Required to handle STARTTLS command.
	tlsConfig *tls.Config

	// onClose is a function that will be called when session is ended.
	onClose func(*Event)

	// conn is a current TCP connection.
	conn net.Conn

	// r is a connection reader.
	r *bufio.Reader

	// w is a connection writer.
	w *bufio.Writer

	// scanner is a connection reader scanner.
	scanner *bufio.Scanner

	// rw is a session log.
	log *bytes.Buffer

	// state is a current state of session.
	state int

	// data stores args passed to the corresponding SMTP commands during
	// a session.
	data *Data
}

// handleConn creates new SMTP session and handles connection with it.
func handleConn(ctx context.Context, conn net.Conn, opts options) error {
	var buf bytes.Buffer

	r := bufio.NewReader(io.TeeReader(conn, &buf))
	w := bufio.NewWriter(io.MultiWriter(conn, &buf))
	scanner := bufio.NewScanner(r)

	sess := &session{
		messages:  opts.messages,
		tlsConfig: opts.tlsConfig,
		onClose:   opts.onClose,
		conn:      conn,
		r:         r,
		w:         w,
		scanner:   scanner,
		log:       &buf,
		state:     stateHelo,
		data: &Data{
			RcptTo: make([]string, 0),
		},
	}

	return sess.start(ctx)
}

func (s *session) start(ctx context.Context) error {
	start := time.Now()

	if s.onClose != nil {
		defer func() {
			_, secure := s.conn.(*tls.Conn)

			s.onClose(&Event{
				RemoteAddr: s.conn.RemoteAddr(),
				Log:        s.log.Bytes(),
				Data:       s.data,
				Secure:     secure,
				ReceivedAt: start,
			})
		}()
	}

	if err := s.greet(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			if !s.scanner.Scan() {
				return s.scanner.Err()
			}

			line := s.scanner.Text()

			if err := s.handle(line); err == ErrQuit {
				return nil
			} else if err != nil {
				return err
			}
		}
	}
}

func (s *session) handle(line string) error {

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

//
// Helpers
//

func (s *session) parseCmd(line string) (string, string) {
	parts := strings.SplitN(line, " ", 2)

	cmd := strings.ToUpper(parts[0])
	args := ""

	if len(parts) > 1 {
		args = parts[1]
	}

	return cmd, args
}

func (s *session) writeLine(line string) error {
	if _, err := s.w.WriteString(line + "\r\n"); err != nil {
		return err
	}
	return s.w.Flush()
}

func (s *session) badSequenceError() error {
	return s.writeLine("503 Bad sequence of commands")
}

func (s *session) commandError() error {
	return s.writeLine("502 Command is not implemented")
}

func (s *session) greet() error {
	return s.writeLine(fmt.Sprintf("220 %s SMTP Server ready", s.messages.Greet))
}

//
// Commands handlers
//

// HELO
func (s *session) handleHelo(args string) error {
	s.data.Helo = args
	s.state = stateMailFrom
	return s.writeLine("250 Hello")
}

// NOOP
func (s *session) handleNoop(args string) error {
	return s.writeLine("250 OK")
}

// EHLO
func (s *session) handleEhlo(args string) error {
	s.data.Ehlo = args
	s.state = stateMailFrom

	if err := s.writeLine(fmt.Sprintf("250-%s", s.messages.Ehlo)); err != nil {
		return err
	}

	if s.tlsConfig != nil {
		if err := s.writeLine("250-STARTTLS"); err != nil {
			return err
		}
	}

	return s.writeLine("250 HELO")
}

// STARTTLS
func (s *session) handleStartTLS(args string) error {

	if s.tlsConfig == nil {
		s.writeLine("502 Command is not implemented")
		return nil
	}

	if err := s.writeLine("220 Ready to start TLS"); err != nil {
		return err
	}

	conn := tls.Server(s.conn, s.tlsConfig)

	if err := conn.Handshake(); err != nil {
		return err
	}

	s.conn = net.Conn(conn)

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, s.log); err != nil {
		return err
	}

	s.r = bufio.NewReader(io.TeeReader(conn, &buf))
	s.w = bufio.NewWriter(io.MultiWriter(conn, &buf))
	s.scanner = bufio.NewScanner(s.r)
	s.log = &buf
	s.state = stateHelo

	return nil
}

// MAIL FROM
func (s *session) handleMailFrom(args string) error {
	s.data.MailFrom = addrRegexp.ReplaceAllString(args, "$2")
	s.state = stateRcptTo
	return s.writeLine("250 OK")
}

// RCPT TO
func (s *session) handleRcptTo(args string) error {
	s.data.RcptTo = append(s.data.RcptTo, addrRegexp.ReplaceAllString(args, "$2"))
	s.state = stateRcptTo
	return s.writeLine("250 OK")
}

// DATA
func (s *session) handleData(args string) error {
	if args == "" && s.state != stateData {
		s.state = stateData
		return s.writeLine("354 Send data")
	} else if args != "." {
		s.data.Data += args + "\n"
		return nil
	}

	s.state = stateMailFrom
	return s.writeLine("250 OK")
}

// QUIT
func (s *session) handleQuit(args string) error {
	_ = s.writeLine("221 OK")
	return ErrQuit
}
