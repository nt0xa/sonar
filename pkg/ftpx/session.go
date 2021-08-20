package ftpx

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

var (
	ErrClose = errors.New("connection closed")
)

// Data stores args passed to the corresponding FTP commands.
type Data struct {
	User string `structs:"user"`
	Pass string `structs:"pass"`

	Type string `structs:"type"`

	Pasv string `structs:"pasv"`
	Epsv string `structs:"epsv"`

	Port string `structs:"port"`
	Eprt string `structs:"eprt"`

	Retr string `structs:"retr"`
}

// Event represents FTP event.
type Event struct {
	// RemoteAddre is remote IP address.
	RemoteAddr net.Addr

	// Log is a full session log.
	Log []byte

	// Data stores args passed to the corresponding FTP commands during a session.
	Data Data

	// Secure shows connection was secure (with TLS).
	Secure bool

	// ReceivedAt is a session start time.
	ReceivedAt time.Time
}

// Msg is used to store provided command responses.
type Msg struct {
	// Greet is server greet message.
	Greet string
}

// session contains all data required to handle FTP session.
type session struct {
	// messages stores provided command responses.
	messages Msg

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

	// data stores args passed to the corresponding FTP commands during
	// a session.
	data Data
}

// handleConn creates new FTP session and handles connection with it.
func handleConn(ctx context.Context, conn net.Conn, opts options) error {
	var buf bytes.Buffer

	r := bufio.NewReader(io.TeeReader(conn, &buf))
	w := bufio.NewWriter(io.MultiWriter(conn, &buf))
	scanner := bufio.NewScanner(r)

	sess := &session{
		messages: opts.messages,
		onClose:  opts.onClose,
		conn:     conn,
		r:        r,
		w:        w,
		scanner:  scanner,
		log:      &buf,
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

			if err := s.handle(line); err == ErrClose {
				return nil
			} else if err != nil {
				return err
			}
		}
	}
}

func (s *session) handle(line string) error {
	cmd, args := s.parseCmd(line)

	switch cmd {
	case "USER":
		return s.handleUser(args)
	case "PASS":
		return s.handlePass(args)
	case "TYPE":
		return s.handleType(args)
	case "EPSV":
		return s.handleEpsv(args)
	case "PASV":
		return s.handlePasv(args)
	case "EPRT":
		return s.handleEprt(args)
	case "PORT":
		return s.handlePort(args)
	case "RETR":
		return s.handleRetr(args)
	case "QUIT":
		return s.handleQuit(args)
	default:
		return s.writeLine("500 Unknown command.")
	}
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

func (s *session) greet() error {
	return s.writeLine(fmt.Sprintf("220 %s", s.messages.Greet))
}

//
// Commands handlers
//

// USER
func (s *session) handleUser(args string) error {
	s.data.User = args
	return s.writeLine("331 Please specify the password.")
}

// PASS
func (s *session) handlePass(args string) error {
	s.data.Pass = args
	return s.writeLine("230 Login successful.")
}

// TYPE
func (s *session) handleType(args string) error {
	s.data.Type = args
	switch strings.ToUpper(args) {
	case "I":
		return s.writeLine("200 Switching to Binary mode.")
	case "A":
		return s.writeLine("200 Switching to ASCII mode.")
	}
	return s.writeLine("500 Unrecognised TYPE command.")
}

// EPSV
func (s *session) handleEpsv(args string) error {
	s.data.Epsv = args
	if strings.ToUpper(args) == "ALL" {
		return s.writeLine("200 EPSV ALL ok.")
	}
	// TODO: maybe use valid IP and port port?
	// Example: "229 Entering Extended Passive Mode (|||1337|)"
	// For now just disallow passive mode.
	return s.writeLine("550 Permission denied.")
}

// PASV
func (s *session) handlePasv(args string) error {
	s.data.Pasv = args
	// TODO: maybe use valid IP and port port?
	// Example: "227 Entering Passive Mode (127,0,0,1,57,5)."
	// For now just disallow passive mode.
	return s.writeLine("550 Permission denied.")
}

// EPRT
func (s *session) handleEprt(args string) error {
	s.data.Eprt = args
	return s.writeLine("200 EPRT command successful")
}

// PORT
func (s *session) handlePort(args string) error {
	s.data.Port = args
	return s.writeLine("200 EPRT command successful")
}

// RETR
func (s *session) handleRetr(args string) error {
	s.data.Retr = args
	// Return 451 error instead of "226 Transfer complete." to force client not to wait for
	// active mode data connection.
	s.writeLine("451 Nope.")
	s.writeLine("221 Goodbye.")
	return ErrClose
}

// QUIT
func (s *session) handleQuit(args string) error {
	s.writeLine("221 Goodbye.")
	return ErrClose
}
