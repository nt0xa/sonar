package server

import "net"

type NotifyFunc func(net.Addr, []byte, map[string]interface{})
