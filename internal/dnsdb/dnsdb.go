package dnsdb

import (
	"database/sql"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/bi-zone/sonar/internal/database"
	"github.com/bi-zone/sonar/internal/models"
	"github.com/bi-zone/sonar/internal/utils/pointer"
	"github.com/bi-zone/sonar/internal/utils/slice"
	"github.com/bi-zone/sonar/pkg/dnsutils"
)

// Records searches for DNS records in the database.
type Records struct {
	DB     *database.DB
	Origin string
}

// Get allows handler to implement dnsx.RecordSet interface.
func (h *Records) Get(name string, qtype uint16) ([]dns.RR, error) {
	// Trim origin because domains are stored without it.
	// test1.test2.00b18489.sonar.local -> [test1, test2, 00b18489]
	parts := strings.Split(strings.TrimSuffix(strings.ToLower(name), "."+h.Origin+"."), ".")

	// This can't be user created domain.
	if len(parts) < 2 {
		return nil, nil
	}

	// Get payload subdomain from name, i.e. rightmost part.
	domain := parts[len(parts)-1]

	payload, err := h.DB.PayloadsGetBySubdomain(domain)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	count, err := h.DB.DNSRecordsGetCountByPayloadID(payload.ID)
	if err != nil {
		return nil, err
	} else if count == 0 {
		return nil, nil
	}

	// Build payload subdomain.
	// [test1 test2 0a88a087] -> test1.test2
	subdomain := strings.Join(parts[:len(parts)-1], ".")

	var record *models.DNSRecord

	typ := dnsutils.QtypeString(qtype)

	for _, n := range dnsutils.MakeWildcards(subdomain) {

		// TODO: add db query for multiple names.
		record, err = h.DB.DNSRecordsGetByPayloadNameAndType(payload.ID, n, typ)
		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			return nil, err
		}

		break
	}

	res := make([]dns.RR, 0)

	if record == nil {
		// Return non nil here to stop handlers chain and return empty answer
		// instead of fallback to default records set.
		// This is required in case when you, for example, want that your subdomain only return
		// AAAA record without A record.
		return res, nil
	}

	// Use name here instead of record.Name because record.Name may be wildcard.
	rrs := dnsutils.NewRRs(name, record.Qtype(), record.TTL, record.Values)

	// Build answer based on record "strategy".
	switch record.Strategy {

	// "all" — just return all values.
	case models.DNSStrategyAll:
		res = rrs

	// "round-robin" — return all records but rotate them cyclically.
	case models.DNSStrategyRoundRobin:
		if record.LastAnswer != nil {
			res = rotate(dnsutils.NewRRs(name, record.Qtype(), record.TTL, record.LastAnswer))
		} else {
			res = rrs
		}

	// "rebind" - if time since last request is less then threshold,
	// return next record, else return first record.
	case models.DNSStrategyRebind:
		if record.LastAnswer != nil &&
			record.LastAccessedAt != nil &&
			len(record.LastAnswer) > 0 &&
			time.Now().Sub(*record.LastAccessedAt) < time.Second*3 {
			i := slice.FindIndex(record.Values, record.LastAnswer[0])
			res = []dns.RR{rrs[min(i+1, len(rrs)-1)]}
		} else {
			// Fallback to first record.
			res = rrs[:1]
		}
	}

	// Update last answer and last answer time.
	record.LastAnswer = dnsutils.RRsToStrings(res)
	record.LastAccessedAt = pointer.Time(time.Now())

	if err := h.DB.DNSRecordsUpdate(record); err != nil {
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
