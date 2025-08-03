package geoipx

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/oschwald/geoip2-golang/v2"
)

type DB struct {
	city *geoip2.Reader
	asn  *geoip2.Reader
}

func New(city, asn string) (*DB, error) {
	var (
		err error
		db  DB
	)

	db.city, err = geoip2.Open(city)
	if err != nil {
		return nil, fmt.Errorf("failed to open city database: %w", err)
	}
	if meta := db.city.Metadata(); meta.DatabaseType != "GeoLite2-City" {
		return nil, fmt.Errorf(
			"expected GeoLite2-City database, got %s",
			city,
		)
	}

	db.asn, err = geoip2.Open(asn)
	if err != nil {
		return nil, fmt.Errorf("failed to open ASN database: %w", err)
	}
	if meta := db.asn.Metadata(); meta.DatabaseType != "GeoLite2-ASN" {
		return nil, fmt.Errorf(
			"expected GeoLite2-ASN database, got %s",
			asn,
		)
	}

	return &db, nil
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
