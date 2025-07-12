package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	"github.com/nt0xa/sonar/pkg/logx"
	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cmd.Context(), &cfg)
		},
	}

	root.AddCommand(serve)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, cfg *server.Config) error {
	errChan := make(chan error, 2)

	//
	// Telemetry
	//

	tel, err := telemetry.New(ctx, "sonar", "v0") // TODO: change
	if err != nil {
		return fmt.Errorf("failed to init telemetry: %w", err)
	}

	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	//
	// Logger
	//

	log := slog.New(logx.MultiHandler(
		slog.NewTextHandler(os.Stdout, nil),
		tel.NewLogHandler("sonar"),
	))

	//
	// DB
	//

	db, err := database.New(cfg.DB.DSN, log)
	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
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

	if u, err := db.UsersGetByName("admin"); errors.Is(err, sql.ErrNoRows) {
		// There is no admin yet - create one
		if err := db.UsersCreate(admin); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
	} else if err == nil {
		// Admin user exists - update
		admin.ID = u.ID
		if err := db.UsersUpdate(admin); err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}
	} else {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	//
	// Cache
	//

	cache, err := cache.New(db)
	if err != nil {
		return err
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
		tel,
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
			errChan <- fmt.Errorf("failed to start DNS handler: %w", err)
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
		return fmt.Errorf("failed to init TLS: %w", err)
	}

	go func() {
		err := tls.Start()
		if err != nil {
			errChan <- fmt.Errorf("failed to start TLS: %w", err)
		}
	}()

	tls.Wait()

	tlsConfig, err := tls.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS config: %w", err)
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
			errChan <- fmt.Errorf("failed to start HTTP handler: %w", err)
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
			errChan <- fmt.Errorf("failed to start HTTPS handler: %w", err)
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
			errChan <- fmt.Errorf("failed to start SMTP handler: %w", err)
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
			errChan <- fmt.Errorf("failed to start FTP handler: %w", err)
		}
	}()

	//
	// Modules
	//

	controllers, notifiers, err := server.Modules(
		&cfg.Modules,
		db,
		log,
		tlsConfig,
		actions,
		cfg.Domain,
	)
	if err != nil {
		return err
	}

	// Start controllers
	for _, c := range controllers {
		go func(c server.Controller) {
			if err := c.Start(); err != nil {
				errChan <- fmt.Errorf("failed to start controller %v: %w", c, err)
			}
		}(c)
	}

	// Add notifiers
	for i, n := range notifiers {
		events.AddNotifier(fmt.Sprintf("Notifier %d", i), n)
	}

	// Process events
	func() {
		if err := events.Start(); err != nil {
			errChan <- fmt.Errorf("failed to start events handler: %w", err)
		}
	}()

	// Wait forever
	log.Info("Starting server")
	if err := <-errChan; err != nil {
		return err
	}

	return nil
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
