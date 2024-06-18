package server

import (
	"net"
	"time"

	"github.com/fatih/structs"
	"github.com/russtone/sonar/internal/database/models"
	"github.com/russtone/sonar/pkg/ftpx"
	"github.com/russtone/sonar/pkg/netx"
)

func FTPListenerWrapper(maxBytes int64, idleTimeout time.Duration) func(net.Listener) net.Listener {
	return func(l net.Listener) net.Listener {
		return &netx.TimeoutListener{
			Listener: &netx.MaxBytesListener{
				Listener: l,
				MaxBytes: maxBytes,
			},
			IdleTimeout: idleTimeout,
		}
	}
}

func FTPEvent(e *ftpx.Event) *models.Event {

	type Meta struct {
		Session ftpx.Data `structs:"session"`
		Secure  bool      `structs:"secure"`
	}

	meta := &Meta{
		Session: e.Data,
		Secure:  e.Secure,
	}

	return &models.Event{
		Protocol:   models.ProtoFTP,
		R:          e.R,
		W:          e.W,
		RW:         e.RW,
		Meta:       structs.Map(meta),
		RemoteAddr: e.RemoteAddr.String(),
		ReceivedAt: e.ReceivedAt,
	}
}
