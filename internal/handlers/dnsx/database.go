package dnsx

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/pointer"
	"github.com/bi-zone/sonar/internal/utils/slice"
)

// DatabaseFinder uses database.DB to find DNS records in the database.
type DatabaseFinder struct {
	db     *database.DB
	origin string
}

// DatabaseFinder must implement Finder interface.
var _ Finder = &DatabaseFinder{}

// NewDatabaseFinder returns new instance of DatabaseFinder.
func NewDatabaseFinder(db *database.DB, origin string) *DatabaseFinder {
	return &DatabaseFinder{db, origin}
}

// Find allows DatabaseFinder to implement Finder interface.
func (f *DatabaseFinder) Find(name string, qtype uint16) ([]dns.RR, error) {
	// test1.test2.00b18489.sonar.local -> [test1, test2, 00b18489]
	parts := strings.Split(strings.TrimSuffix(name, "."+f.origin+"."), ".")

	if len(parts) < 2 {
		return nil, nil
	}

	// Get payload subdomain from name, i.e. rightmost part.
	domain := parts[len(parts)-1]

	payload, err := f.db.PayloadsGetBySubdomain(domain)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// Build payload subdomain.
	// [test1 test2 0a88a087] -> test1.test2
	subdomain := strings.Join(parts[:len(parts)-1], ".")

	record, err := f.db.DNSRecordsGetByPayloadNameType(payload.ID, subdomain, qtypeStr(qtype))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var res []dns.RR

	fqdn := fmt.Sprintf("%s.%s.%s.", record.Name, payload.Subdomain, f.origin)
	rrs := NewRRs(fqdn, record.Qtype(), record.TTL, record.Values)

	// Build answer based on record "strategy".
	switch record.Strategy {

	// "all" — just return all values.
	case models.DNSStrategyAll:
		res = rrs

	// "round-robin" — return all records but rotate them cyclically.
	case models.DNSStrategyRoundRobin:
		if record.LastAnswer != nil {
			res = rotate(NewRRs(fqdn, record.Qtype(), record.TTL, record.LastAnswer))
		} else {
			res = rrs
		}

	// "rebind" - if time since last request is less then threshold,
	// return next record, else return first record.
	case models.DNSStrategyRebind:
		if record.LastAnswer != nil &&
			record.LastAccessedAt != nil &&
			len(record.LastAnswer) > 0 &&
			time.Now().UTC().Sub(*record.LastAccessedAt) < time.Second*3 {
			i := slice.FindIndex(record.Values, record.LastAnswer[0])
			res = []dns.RR{rrs[min(i+1, len(rrs)-1)]}
		} else {
			// Fallback to first record.
			res = rrs[:1]
		}
	}

	// Update last answer and last answer time.
	record.LastAnswer = rrsToStrings(res)
	record.LastAccessedAt = pointer.Time(time.Now().UTC())

	if err := f.db.DNSRecordsUpdate(record); err != nil {
		return nil, err
	}

	return res, nil
}

// rotate cyclically shifts records left by 1.
func rotate(rrs []dns.RR) []dns.RR {
	if len(rrs) <= 1 {
		return rrs
	}
	newRRs := make([]dns.RR, len(rrs))
	copy(newRRs, rrs[1:])
	newRRs[len(rrs)-1] = rrs[0]
	return newRRs
}

// min returns minimum of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
