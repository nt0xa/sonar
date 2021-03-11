package netx

import (
	"crypto/tls"
	"net"
)

type Server struct {
	Addr              string
	TLSConfig         *tls.Config
	NotifyStartedFunc func()
	ListenerWrapper   func(net.Listener) net.Listener
	ConnectionHandler func(net.Conn) error
}

func (s *Server) ListenAndServe() error {
	var (
		err      error
		listener net.Listener
	)

	if s.TLSConfig != nil {
		listener, err = tls.Listen("tcp", s.Addr, s.TLSConfig)
	} else {
		listener, err = net.Listen("tcp", s.Addr)
	}

	if err != nil {
		return err
	}

	defer listener.Close()

	l := listener

	if s.ListenerWrapper != nil {
		l = s.ListenerWrapper(l)
	}

	if s.NotifyStartedFunc != nil {
		s.NotifyStartedFunc()
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		go func() {
			// TODO: logging
			defer conn.Close()
			_ = s.ConnectionHandler(conn)
		}()
	}

}
