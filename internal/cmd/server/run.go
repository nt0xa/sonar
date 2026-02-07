package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/nt0xa/sonar/internal/actionsdb"
	"github.com/nt0xa/sonar/internal/cache"
	"github.com/nt0xa/sonar/internal/database"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/nt0xa/sonar/pkg/ftpx"
	"github.com/nt0xa/sonar/pkg/geoipx"
	"github.com/nt0xa/sonar/pkg/httpx"
	"github.com/nt0xa/sonar/pkg/logx"
	"github.com/nt0xa/sonar/pkg/smtpx"
	"github.com/nt0xa/sonar/pkg/telemetry"
)

func Run(
	ctx context.Context,
	stdout io.Writer,
	stderr io.Writer,
	dir fs.FS,
	environFunc func() []string,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := LoadConfig(
		dir,
		environFunc,
	)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	errChan := make(chan error, 2)

	//
	// Telemetry
	//

	var tel telemetry.Telemetry

	if cfg.Telemetry.Enabled {
		tel, err = telemetry.New(ctx, "sonar", "v0") // TODO: change
		if err != nil {
			return fmt.Errorf("failed to init telemetry: %w", err)
		}
	} else {
		tel = telemetry.NewNoop()
	}

	defer func() {
		_ = tel.Shutdown(ctx)
	}()

	ctx, span := tel.TraceStart(
		ctx,
		"server.start",
		trace.WithSpanKind(trace.SpanKindInternal),
	)

	//
	// Logger
	//

	log := slog.New(logx.MultiHandler(
		slog.NewTextHandler(stdout, nil),
		tel.NewLogHandler("sonar"),
	))

	//
	// DB
	//

	db, err := database.NewWithDSN(cfg.DB.DSN)
	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	if _, err := database.Migrate(cfg.DB.DSN); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create admin user
	if err := createOrUpdateAdminUser(ctx, cfg, db, tel); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	//
	// Cache
	//

	cache, err := cache.New(ctx, db)
	if err != nil {
		return err
	}

	//
	// Actions
	//

	actions := actionsdb.New(
		db,
		log.With("package", "actions"),
		cfg.Domain,
	)

	//
	// GeoIP
	//

	var gdb *geoipx.DB

	if cfg.GeoIP.Enabled {
		gdb, err = geoipx.New(log.With("package", "geoipx"), cfg.GeoIP.City, cfg.GeoIP.ASN)
		if err != nil {
			return fmt.Errorf("failed to init GeoIP database: %w", err)
		}

		if err := gdb.Watch(ctx); err != nil {
			return fmt.Errorf("failed to watch GeoIP database: %w", err)
		}
	}

	//
	// EventsHandler
	//

	events := NewEventsHandler(
		db,
		gdb,
		log.With("package", "events"),
		tel,
		cache,
		10,
		100,
	)

	//
	// DNS
	//

	var waitDNS sync.WaitGroup

	waitDNS.Add(1)

	dnsHandler := DNSHandler(
		&cfg.DNS,
		db,
		tel,
		cfg.Domain,
		net.ParseIP(cfg.IP),
		emitDNS(events),
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

	tls, err := NewTLS(&cfg.TLS, log, cfg.Domain, dnsHandler)
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
			HTTPHandler(
				db,
				tel,
				cfg.Domain,
				emitHTTP(events),
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
			HTTPHandler(
				db,
				tel,
				cfg.Domain,
				emitHTTP(events),
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
			SMTPHandler(
				cfg.Domain,
				log.With("package", "smtpx"),
				tel,
				tlsConfig,
				emitSMTP(events),
			),
			smtpx.ListenerWrapper(SMTPListenerWrapper(1<<20, time.Second*5)), // TODO: change to handler
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
			FTPHandler(
				cfg.Domain,
				log.With("package", "ftpx"),
				tel,
				emitFTP(events),
			),
			ftpx.ListenerWrapper(SMTPListenerWrapper(1<<20, time.Second*5)),
		)

		if err := srv.ListenAndServe(); err != nil {
			errChan <- fmt.Errorf("failed to start FTP handler: %w", err)
		}
	}()

	//
	// Modules
	//

	controllers, notifiers, err := Modules(
		&cfg.Modules,
		db,
		log,
		tel,
		tlsConfig,
		actions,
		cfg.Domain,
	)
	if err != nil {
		return err
	}

	// Start controllers
	for _, c := range controllers {
		go func(c Controller) {
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
	if err := events.Start(); err != nil {
		return fmt.Errorf("failed to start events handler: %w", err)
	}

	// Wait forever
	log.Info("Starting server")
	span.End()

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

func createOrUpdateAdminUser(
	ctx context.Context,
	cfg *Config,
	db *database.DB,
	tel telemetry.Telemetry,
) error {
	ctx, span := tel.TraceStart(ctx, "db.init.admin")
	defer span.End()

	if u, err := db.UsersGetByName(ctx, "admin"); errors.Is(err, database.ErrNoRows) {
		// There is no admin yet - create one
		if _, err := db.UsersCreate(ctx, database.UsersCreateParams{
			Name:       "admin",
			IsAdmin:    true,
			CreatedBy:  nil,
			TelegramID: &cfg.Modules.Telegram.Admin,
			APIToken:   &cfg.Modules.API.Admin,
			LarkID:     &cfg.Modules.Lark.Admin,
			SlackID:    &cfg.Modules.Slack.Admin,
		}); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
	} else if err == nil {
		// Admin user exists - update
		if _, err := db.UsersUpdate(ctx, database.UsersUpdateParams{
			ID:         u.ID,
			Name:       "admin",
			IsAdmin:    true,
			CreatedBy:  nil,
			TelegramID: &cfg.Modules.Telegram.Admin,
			APIToken:   &cfg.Modules.API.Admin,
			LarkID:     &cfg.Modules.Lark.Admin,
			SlackID:    &cfg.Modules.Slack.Admin,
		}); err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}
	} else {
		return fmt.Errorf("failed to get admin user: %w", err)
	}

	return nil
}
