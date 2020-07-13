package handlers

import "net"

type NotifyRequestFunc func(remoteAddr net.Addr, data []byte, meta map[string]interface{})
type NotifyStartedFunc func()
