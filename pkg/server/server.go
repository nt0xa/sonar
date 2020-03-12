package server

import (
	"net"
)

type NotifyRequestFunc func(remoteAddr net.Addr, data []byte, meta map[string]interface{})

type RequestNotifier interface {
	Notify(net.Addr, []byte, map[string]interface{})
}

type NotifyStartedFunc func()

type StartNotifier interface {
	Notify()
}
