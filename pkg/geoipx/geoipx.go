package geoipx

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/oschwald/geoip2-golang/v2"
)

type DB struct {
	log *slog.Logger

	cityPath string
	asnPath  string

	mu   sync.RWMutex
	city *geoip2.Reader
	asn  *geoip2.Reader

	watcher *fsnotify.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
}

func New(log *slog.Logger, city, asn string) (*DB, error) {
	db := &DB{
		log:      log,
		cityPath: city,
		asnPath:  asn,
	}

	if err := db.loadReaders(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) loadReaders() error {
	city, err := geoip2.Open(db.cityPath)
	if err != nil {
		return fmt.Errorf("failed to open city database: %w", err)
	}
	if meta := city.Metadata(); meta.DatabaseType != "GeoLite2-City" {
		_ = city.Close()
		return fmt.Errorf(
			"expected GeoLite2-City database, got %s",
			db.cityPath,
		)
	}

	asn, err := geoip2.Open(db.asnPath)
	if err != nil {
		_ = city.Close()
		return fmt.Errorf("failed to open ASN database: %w", err)
	}
	if meta := asn.Metadata(); meta.DatabaseType != "GeoLite2-ASN" {
		_ = city.Close()
		_ = asn.Close()
		return fmt.Errorf(
			"expected GeoLite2-ASN database, got %s",
			db.asnPath,
		)
	}

	db.updateReaders(city, asn)

	return nil
}

func (db *DB) updateReaders(city, asn *geoip2.Reader) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.city != nil {
		db.city.Close()
	}

	if db.asn != nil {
		db.asn.Close()
	}

	db.city = city
	db.asn = asn
}

func (db *DB) Watch(ctx context.Context) error {
	if db.watcher != nil {
		return fmt.Errorf("file watcher already started")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	if err := watcher.Add(db.cityPath); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("failed to watch city file: %w", err)
	}

	if err := watcher.Add(db.asnPath); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("failed to watch ASN file: %w", err)
	}

	db.watcher = watcher
	db.ctx, db.cancel = context.WithCancel(ctx)

	go db.watchFiles()

	db.log.Info("Started GeoIP file watcher",
		"cityFile", db.cityPath,
		"asnFile", db.asnPath,
	)

	return nil
}

func (db *DB) Stop() {
	if db.cancel != nil {
		db.cancel()
	}
	if db.watcher != nil {
		db.watcher.Close()
		db.watcher = nil
	}
}

func (db *DB) watchFiles() {
	defer db.watcher.Close()

	for {
		select {
		case <-db.ctx.Done():
			return
		case event, ok := <-db.watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				if event.Name == db.cityPath || event.Name == db.asnPath {
					if err := db.loadReaders(); err != nil {
						db.log.Error("Failed to reload GeoIP databases", "err", err)
					} else {
						db.log.Info("Reloaded GeoIP databases after file change", "file", event.Name)
					}
				}
			}
		case err, ok := <-db.watcher.Errors:
			if !ok {
				return
			}
			db.log.Error("File watcher error", "err", err)
		}
	}
}

func (db *DB) Close() error {
	db.Stop()

	db.mu.Lock()
	defer db.mu.Unlock()

	var errs []error
	if db.city != nil {
		if err := db.city.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close city database: %w", err))
		}
	}
	if db.asn != nil {
		if err := db.asn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close ASN database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}

type Info struct {
	City         string      `json:"city"`
	Country      CountryInfo `json:"country"`
	Subdivisions []string    `json:"subdivisions"`
	ASN          ASNInfo     `json:"asn"`
}

type CountryInfo struct {
	Name      string `json:"name"`
	ISOCode   string `json:"isoCode"`
	FlagEmoji string `json:"flagEmoji"`
}

type ASNInfo struct {
	Number uint   `json:"number"`
	Org    string `json:"org"`
}

func (db *DB) Lookup(ip netip.Addr) (*Info, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	city, err := db.city.City(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup city: %w", err)
	}

	asn, err := db.asn.ASN(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup ASN: %w", err)
	}

	subdivisions := make([]string, 0)
	for _, s := range city.Subdivisions {
		subdivisions = append(subdivisions, s.Names.English)
	}

	return &Info{
		City:         city.City.Names.English,
		Subdivisions: subdivisions,
		Country: CountryInfo{
			Name:      city.Country.Names.English,
			ISOCode:   city.Country.ISOCode,
			FlagEmoji: flagEmoji(city.Country.ISOCode),
		},
		ASN: ASNInfo{
			Number: asn.AutonomousSystemNumber,
			Org:    asn.AutonomousSystemOrganization,
		},
	}, nil
}

func flagEmoji(countryCode string) string {
	countryCode = strings.ToUpper(countryCode)
	if len(countryCode) != 2 {
		return "" // Invalid code
	}
	runes := []rune{}
	for _, c := range countryCode {
		if c < 'A' || c > 'Z' {
			return "" // Invalid character
		}
		runes = append(runes, 0x1F1E6+(c-'A'))
	}
	return string(runes)
}
