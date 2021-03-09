package main

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules"
	"github.com/bi-zone/sonar/internal/utils/slice"
)

var (
	subdomainRegexp = regexp.MustCompile("[a-f0-9]{8}")
)

func ProcessEvents(log *logrus.Logger, events <-chan models.Event, db *database.DB, ns []modules.Notifier) error {
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

			protocol := strings.ToLower(e.Protocol)

			// Change "https" to "http" because there is only
			// one category for both
			if protocol == "https" {
				protocol = "http"
			}

			// Skip if current event protocol is muted for payload.
			if !slice.StringsContains(p.NotifyProtocols, protocol) {
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

func AddProtoEvent(proto string, events chan<- models.Event) func(net.Addr, []byte, map[string]interface{}) {
	return func(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {

		events <- models.Event{
			Protocol:   proto,
			Data:       string(data),
			RawData:    data,
			Meta:       meta,
			RemoteAddr: remoteAddr,
			ReceivedAt: time.Now(),
		}
	}
}
