package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/cache"
	"github.com/nt0xa/sonar/internal/cmd/server"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/internal/database/models"
	"github.com/nt0xa/sonar/internal/utils"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/smtpx"
)

func init() {
	validation.ErrorTag = "mapstructure"
}

func main() {
	var (
		cfg     server.Config
		cfgFile string
	)

	root := &cobra.Command{
		Use:   "server",
		Short: "CLI for sonar server",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if err := initConfig(cfgFile, &cfg); err != nil {
				return fmt.Errorf("fail to init config: %w", err)
			}

			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

	serve := &cobra.Command{
		Use:   "serve",
		Short: "start the server",
		Run: func(cmd *cobra.Command, args []string) {
			serve(&cfg)
		},
	}

	root.AddCommand(serve)

	root.Execute()
}

func serve(cfg *server.Config) {
	//
	// Logger
	//

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})

	//
	// DB
	//

	db, err := database.New(cfg.DB.DSN, log)
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
			LarkUserID: cfg.Modules.Lark.Admin,
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
	// Cache
	//

	cache, err := cache.New(db)
	if err != nil {
		log.Fatal(err)
	}

	//
	// Actions
	//

	actions := actionsdb.New(db, log, cfg.Domain)

	//
	// EventsHandler
	//

	events := server.NewEventsHandler(db, cache, 10, 100 /* TODO: from config */)

	//
	// DNS
	//

	var waitDNS sync.WaitGroup

	waitDNS.Add(1)

	dnsHandler := server.DNSHandler(
		&cfg.DNS,
		db,
		cfg.Domain,
		net.ParseIP(cfg.IP),
		func(e *dnsx.Event) {
			events.Emit(server.DNSEvent(e))
		},
	)

	go func() {
		srv := dnsx.New(
			":5053",
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
			smtpx.ListenerWrapper(server.SMTPListenerWrapper(1<<20, time.Second*5)),
			smtpx.Messages(smtpx.Msg{Greet: cfg.Domain, Ehlo: cfg.Domain}),
			smtpx.OnClose(func(e *smtpx.Event) {
				events.Emit(server.SMTPEvent(e))
			}),
			smtpx.TLSConfig(tlsConfig, false),
		)

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed start SMTP handler: %s", err.Error())
		}
	}()

	//
	// FTP
	//

	go func() {
		// Pass TLS config to be able to handle "STARTTLS" command.
		srv := ftpx.New(
			":21",
			ftpx.ListenerWrapper(server.SMTPListenerWrapper(1<<20, time.Second*5)),
			ftpx.Messages(ftpx.Msg{Greet: fmt.Sprintf("%s Server ready", cfg.Domain)}),
			ftpx.OnClose(func(e *ftpx.Event) {
				events.Emit(server.FTPEvent(e))
			}),
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

func initConfig(cfgFile string, cfg *server.Config) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("sonar")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for _, key := range utils.StructKeys(*cfg, "mapstructure") {
		viper.BindEnv(key)
	}

	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, new(viper.ConfigFileNotFoundError)) {
			return err
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}
