package main

import (
	"database/sql"
	"net"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/cmd/server"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules"
	"github.com/bi-zone/sonar/internal/tls"
	"github.com/bi-zone/sonar/pkg/dnsx"
	"github.com/bi-zone/sonar/pkg/httpx"
	"github.com/bi-zone/sonar/pkg/smtpx"
)

func main() {

	//
	// Logger
	//

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})

	//
	// Config
	//

	var cfg Config

	if err := envconfig.Process("sonar", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	//
	// DB
	//

	db, err := database.New(&cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Create admin user
	admin := &models.User{
		Name:      "admin",
		IsAdmin:   true,
		CreatedBy: nil,
		Params: models.UserParams{
			TelegramID: cfg.Modules.Telegram.Admin,
			APIToken:   cfg.Modules.API.Admin,
		},
	}

	if u, err := db.UsersGetByName("admin"); err == sql.ErrNoRows {
		// There is no admin yet - create one
		if err := db.UsersCreate(admin); err != nil {
			log.Fatal(err)
		}
	} else if err == nil {
		// Admin user exists - update
		admin.ID = u.ID
		if err := db.UsersUpdate(admin); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	//
	// Actions
	//

	actions := actionsdb.New(db, log, cfg.Domain)

	//
	// Events
	//

	events := make(chan models.Event, 100)
	defer close(events)

	//
	// DNS
	//

	var waitDNS sync.Mutex

	dnsHandler := server.DNSHandler(
		db,
		cfg.Domain,
		net.ParseIP(cfg.IP),
		AddProtoEvent("DNS", events),
	)

	go func() {
		srv := dnsx.New(
			":53",
			dnsHandler,
			dnsx.NotifyStartedFunc(waitDNS.Unlock),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS handler: %v", err.Error())
		}
	}()

	//
	// TLS
	//

	// Wait for DNS server to start, because we need to
	// use it as DNS challenge provider for Let's Encrypt
	waitDNS.Lock()

	tls, err := tls.New(&cfg.TLS, log, cfg.Domain, dnsHandler)
	if err != nil {
		log.Fatalf("Failed to init TLS: %v", err)
	}

	go func() {
		err := tls.Start()
		if err != nil {
			log.Fatalf("Failed to start TLS: %v", err)
		}
	}()

	tls.Wait()

	tlsConfig, err := tls.GetConfig()
	if err != nil {
		log.Fatalf("Failed to get TLS config: %v", err)
	}

	//
	// HTTP
	//

	go func() {
		srv := httpx.New(
			":80",
			server.HTTPHandler(AddProtoEvent("HTTP", events)),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start HTTP handler: %s", err.Error())
		}

	}()

	//
	// HTTPS
	//

	go func() {
		srv := httpx.New(
			":443",
			server.HTTPHandler(AddProtoEvent("HTTPS", events)),
			httpx.TLSConfig(tlsConfig),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start HTTPS handler: %s", err.Error())
		}
	}()

	//
	// SMTP
	//

	go func() {
		// Pass TLS config to be able to handle "STARTTLS" command.
		srv := smtpx.New(
			":25",
			server.SMTPListenerWrapper(1<<20, time.Second*5),
			server.SMTPSession(cfg.Domain, tlsConfig, AddProtoEvent("SMTP", events)),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start SMTP handler: %s", err.Error())
		}
	}()

	//
	// Modules
	//

	controllers, notifiers, err := modules.Init(&cfg.Modules, db, log, tlsConfig, actions, cfg.Domain)
	if err != nil {
		log.Fatal(err)
	}

	// Start controllers
	for _, c := range controllers {
		go func(c modules.Controller) {
			if err := c.Start(); err != nil {
				log.Fatal(err)
			}
		}(c)
	}

	// Process events
	go ProcessEvents(log, events, db, notifiers)

	// Wait forever
	select {}
}
