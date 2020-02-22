package server

import "net"

type HandlerFunc func(remoteAddr net.Addr, proto string, data []byte)
