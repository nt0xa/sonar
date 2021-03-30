package main

import (
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/actionsdb"
	"github.com/bi-zone/sonar/internal/cmd/server"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
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

	var cfg server.Config

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
	// EventsHandler
	//

	events := server.NewEventsHandler(db, 100)

	//
	// DNS
	//

	var waitDNS sync.WaitGroup

	waitDNS.Add(1)

	dnsHandler := server.DNSHandler(
		db,
		cfg.Domain,
		net.ParseIP(cfg.IP),
		func(e *dnsx.Event) {
			events.Emit(server.DNSEvent(e))
		},
	)

	go func() {
		srv := dnsx.New(
			":53",
			dnsHandler,
			dnsx.NotifyStartedFunc(waitDNS.Done),
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
	waitDNS.Wait()

	tls, err := server.NewTLS(&cfg.TLS, log, cfg.Domain, dnsHandler)
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
			server.HTTPHandler(
				db,
				cfg.Domain,
				func(e *httpx.Event) {
					events.Emit(server.HTTPEvent(e))
				},
			),
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
			server.HTTPHandler(
				db,
				cfg.Domain,
				func(e *httpx.Event) {
					events.Emit(server.HTTPEvent(e))
				},
			),
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
			server.SMTPSession(cfg.Domain, tlsConfig,
				func(e *smtpx.Event) {
					events.Emit(server.SMTPEvent(e))
				},
			),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start SMTP handler: %s", err.Error())
		}
	}()

	//
	// Modules
	//

	controllers, notifiers, err := server.Modules(&cfg.Modules, db, log, tlsConfig, actions, cfg.Domain)
	if err != nil {
		log.Fatal(err)
	}

	// Start controllers
	for _, c := range controllers {
		go func(c server.Controller) {
			if err := c.Start(); err != nil {
				log.Fatal(err)
			}
		}(c)
	}

	// Add notifiers
	for i, n := range notifiers {
		events.AddNotifier(fmt.Sprintf("%d", i), n)
	}

	// Process events
	go events.Start()

	// Wait forever
	select {}
}
