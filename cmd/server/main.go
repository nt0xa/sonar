package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	legolog "github.com/go-acme/lego/v3/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/database/migrations"
	"github.com/bi-zone/sonar/internal/notifier"
	"github.com/bi-zone/sonar/pkg/certmanager"
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

	db, err := database.New(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	if err := migrations.Up(cfg.DB); err != nil {
		log.Fatal(err)
	}

	// Create admin user
	if _, err := db.UsersGetByName("admin"); err == sql.ErrNoRows {
		// There is no admin yet
		u := &database.User{Name: "admin"}
		if err := db.UsersCreate(u); err != nil {
			log.Fatal(err)
		}
	}

	//
	// Interfaces
	//

	is, err := GetEnabledInterfaces(&cfg.Interface, db, cfg.Domain)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range is {
		go func() {
			if err := i.Start(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	//
	// Notifiers
	//

	ns, err := GetEnabledNotifiers(&cfg.Notifier)
	if err != nil {
		log.Fatal(err)
	}

	//
	// Events
	//

	events := make(chan notifier.Event, 100)
	defer close(events)
	handlerFunc := AddEvent(events)
	go ProcessEvents(events, db, ns)

	//
	// DNS
	//

	var (
		dnsProvider *dns.Server
		dnsStarted  sync.WaitGroup
	)

	dnsStarted.Add(1)

	go func() {
		srv := dns.New(":53", cfg.Domain, net.ParseIP(cfg.IP),
			dns.NotifyRequestFunc(AddProtoEvent("DNS", events)),
			dns.NotifyStartedFunc(func() {
				dnsStarted.Done()
			}))

		dnsProvider = srv

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start DNS handler: %s", err.Error())
		}

	}()

	//
	// TLS
	//

	var tlsConfig *tls.Config

	switch cfg.TLS.Type {
	case "custom":
		cert, err := tls.LoadX509KeyPair(cfg.TLS.Custom.Cert, cfg.TLS.Custom.Key)
		if err != nil {
			log.Fatal(err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

	case "letsencrypt":
		legolog.Logger = log
		domains := []string{
			cfg.Domain,                      // domain itseld
			fmt.Sprintf("*.%s", cfg.Domain), // wildcard
		}

		// Wait for DNS server to start
		dnsStarted.Wait()

		certmgr, err := certmanager.New(cfg.TLS.LetsEncrypt.Directory,
			cfg.TLS.LetsEncrypt.Email, dnsProvider)
		if err != nil {
			log.Fatal(err)
		}

		// Try to load existing certificate and key
		cert, err := certmgr.LoadX509KeyPair(domains)
		if err != nil {

			// Obtain new certificate
			newCert, err := certmgr.Obtain(domains)
			if err != nil {
				log.Fatalf("Fail to obtain new certificate: %v", err)
			}

			cert = newCert
		} else {

			// Try to renew cert
			newCert, err := certmgr.Renew(domains)
			if err != nil {
				log.Fatalf("Fail to renew cert: %v", err)
			}

			if newCert != nil {
				cert = newCert
			}
		}

		kpr := &keypairReloader{cert: cert}

		// Auto renew certificates
		ticker := time.NewTicker(12 * time.Hour)
		go func() {
			for range ticker.C {
				renewCertificate(certmgr, kpr, domains)
			}
		}()

		// Load TLS certificate
		tlsConfig = &tls.Config{
			GetCertificate: kpr.GetCertificateFunc(),
		}
	}

	//
	// HTTP
	//

	go func() {
		srv := http.New(":80",
			http.NotifyRequestFunc(AddProtoEvent("HTTP", events)))

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
		srv := smtp.NewServer(":25", cfg.Domain, handlerFunc, smtp.TLSConfig(tlsConfig))

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start SMTP handler: %s", err.Error())
		}
	}()

	// Wait forever
	select {}
}
