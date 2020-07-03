package main

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules"
	"github.com/bi-zone/sonar/internal/utils/slice"
	"github.com/bi-zone/sonar/pkg/server"
)

var (
	subdomainRegexp = regexp.MustCompile("[a-f0-9]{8}")
)

func ProcessEvents(events <-chan models.Event, db *database.DB, ns []modules.Notifier) error {
	for e := range events {

		seen := make(map[string]struct{})

		matches := subdomainRegexp.FindAllSubmatch(e.RawData, -1)
		if len(matches) == 0 {
			continue
		}

		for _, m := range matches {
			d := string(m[0])
			if _, ok := seen[d]; !ok {
				seen[d] = struct{}{}
			} else {
				continue
			}

			p, err := db.PayloadsGetBySubdomain(d)
			if err != nil {
				// TODO: as argument
				log.Println(err)
				continue
			}

			// Skip if current event protocol is muted for payload.
			if !slice.StringsContains(p.NotifyProtocols, strings.ToLower(e.Protocol)) {
				continue
			}

			u, err := db.UsersGetByID(p.UserID)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, n := range ns {
				if err := n.Notify(&e, u, p); err != nil {
					log.Println(err)
					continue
				}
			}

		}
	}

	return nil
}

func AddProtoEvent(proto string, events chan<- models.Event) server.NotifyRequestFunc {
	return func(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {

		events <- models.Event{
			Protocol:   proto,
			Data:       string(data),
			RawData:    data,
			RemoteAddr: remoteAddr,
			ReceivedAt: time.Now(),
		}
	}
}
