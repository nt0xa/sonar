package main

// TODO: make interface

import (
	"fmt"
	"net"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/notifier"
	"github.com/bi-zone/sonar/internal/notifier/telegram"
	"github.com/bi-zone/sonar/internal/utils"
	"github.com/bi-zone/sonar/pkg/server"
)

var (
	subdomainRegexp = regexp.MustCompile("[a-f0-9]{8}")
)

type NotifierConfig struct {
	Enabled  []string        `json:"enabled"`
	Telegram telegram.Config `json:"telegram"`
}

func (c NotifierConfig) Validate() error {
	rules := make([]*validation.FieldRules, 0)
	rules = append(rules, validation.Field(&c.Enabled, validation.Each(validation.In("telegram"))))

	for _, name := range c.Enabled {
		switch name {

		case "telegram":
			rules = append(rules, validation.Field(&c.Telegram))
		}
	}

	return validation.ValidateStruct(&c, rules...)
}

func GetEnabledNotifiers(cfg *NotifierConfig) ([]notifier.Notifier, error) {
	ns := make([]notifier.Notifier, 0)

	var (
		n   notifier.Notifier
		err error
	)

	for _, name := range cfg.Enabled {
		switch name {
		case "telegram":
			n, err = telegram.New(&cfg.Telegram)
		default:
			return nil, fmt.Errorf("unknown notifier %v", name)
		}

		if err != nil {
			return nil, fmt.Errorf("fail to create notifier %v: %w", name, err)
		}

		ns = append(ns, n)
	}

	return ns, nil
}

func ProcessEvents(events <-chan notifier.Event, db *database.DB, ns []notifier.Notifier) error {
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

			for _, n := range ns {
				if err := n.Notify(&e, p); err != nil {
					log.Println(err)
					continue
				}
			}

		}
	}

	return nil
}

// TODO: remove after refactoring
func AddEvent(events chan<- notifier.Event) server.HandlerFunc {
	return func(remoteAddr net.Addr, proto string, data []byte) {

		s := string(data)

		if !utils.StringPrintable(s) {
			s = utils.HexDump(data)
		}

		events <- notifier.Event{
			Protocol:   proto,
			Data:       s,
			RawData:    data,
			RemoteAddr: remoteAddr,
			ReceivedAt: time.Now(),
		}
	}
}

// TODO: use for all after refactoring
func AddProtoEvent(proto string, events chan<- notifier.Event) server.NotifyRequestFunc {
	return func(remoteAddr net.Addr, data []byte, meta map[string]interface{}) {

		s := string(data)

		if !utils.StringPrintable(s) {
			s = utils.HexDump(data)
		}

		events <- notifier.Event{
			Protocol:   proto,
			Data:       s,
			RawData:    data,
			RemoteAddr: remoteAddr,
			ReceivedAt: time.Now(),
		}
	}
}
