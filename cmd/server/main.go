package main

import (
	"database/sql"
	"encoding/json"
	"net"
	"sync"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/actions"
	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/dnsmgr"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/modules"
	"github.com/bi-zone/sonar/internal/tls"
	"github.com/bi-zone/sonar/pkg/server/dns"
	"github.com/bi-zone/sonar/pkg/server/http"
	"github.com/bi-zone/sonar/pkg/server/smtp"
)

var (
	log *logrus.Logger
)

func main() {

	//
	// Logger
	//

	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})

	//
	// Config
	//

	var cfg Config

	if err := envconfig.Process("sonar", &cfg); err != nil {
		log.Fatal(err.Error())
	}

	prettyCfg, _ := json.MarshalIndent(&cfg, "", "  ")
	log.Infof("Config:\n%+v", string(prettyCfg))

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

	actions := actions.New(db, log, cfg.Domain)

	//
	// Events
	//

	events := make(chan models.Event, 100)
	defer close(events)

	//
	// DNS
	//

	dnsmgr, err := dnsmgr.New(cfg.Domain, net.ParseIP(cfg.IP), "[a-f0-9]{8}", db)
	if err != nil {
		log.Fatalf("Fail to create DNS manager: %v", err)
	}

	var dnsStarted sync.WaitGroup

	dnsStarted.Add(1)

	go func() {
		srv := dns.New(":53", dnsmgr.HandleFunc,
			dns.NotifyRequestFunc(AddProtoEvent("DNS", events)),
			dns.NotifyStartedFunc(func() {
				dnsStarted.Done()
			}),
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
	dnsStarted.Wait()

	tls, err := tls.New(&cfg.TLS, log, cfg.Domain, dnsmgr)
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
		srv := http.New(":80", http.NotifyRequestFunc(AddProtoEvent("HTTP", events)))

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start HTTP handler: %s", err.Error())
		}

	}()

	//
	// HTTPS
	//

	go func() {
		srv := http.New(":443",
			http.NotifyRequestFunc(AddProtoEvent("HTTPS", events)),
			http.TLSConfig(tlsConfig))

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start HTTPS handler: %s", err.Error())
		}
	}()

	//
	// SMTP
	//

	go func() {
		// Pass TLS config to be able to handle "STARTTLS" command.
		srv := smtp.New(":25", cfg.Domain,
			smtp.NotifyRequestFunc(AddProtoEvent("SMTP", events)),
			smtp.TLSConfig(tlsConfig),
			smtp.StartTLS(true))

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
	go ProcessEvents(events, db, notifiers)

	// Wait forever
	select {}
}
