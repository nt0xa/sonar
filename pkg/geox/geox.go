package geox

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/oschwald/geoip2-golang/v2"
)

type GeoDB struct {
	city *geoip2.Reader
	asn  *geoip2.Reader
}

func New(city, asn string) (*GeoDB, error) {
	var (
		err error
		gdb GeoDB
	)

	gdb.city, err = geoip2.Open(city)
	if err != nil {
		return nil, fmt.Errorf("failed to open city database: %w", err)
	}
	if meta := gdb.city.Metadata(); meta.DatabaseType != "GeoLite2-City" {
		return nil, fmt.Errorf(
			"expected GeoLite2-City database, got %s",
			city,
		)
	}

	gdb.asn, err = geoip2.Open(asn)
	if err != nil {
		return nil, fmt.Errorf("failed to open ASN database: %w", err)
	}
	if meta := gdb.asn.Metadata(); meta.DatabaseType != "GeoLite2-ASN" {
		return nil, fmt.Errorf(
			"expected GeoLite2-ASN database, got %s",
			asn,
		)
	}

	return &gdb, nil
}

type Info struct {
	City         string
	Country      Country
	Subdivisions []string
	ASN          ASN
}

type Country struct {
	Name    string
	ISOCode string
	Flag    string
}

type ASN struct {
	Number uint
	Org    string
}

func (gdb *GeoDB) Lookup(ip netip.Addr) (*Info, error) {
	city, err := gdb.city.City(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup city: %w", err)
	}

	asn, err := gdb.asn.ASN(ip)
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
		Country: Country{
			Name:    city.Country.Names.English,
			ISOCode: city.Country.ISOCode,
			Flag:    flagEmoji(city.Country.ISOCode),
		},
		ASN: ASN{
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
