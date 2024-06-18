package models_test

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"

	"github.com/nt0xa/sonar/internal/database/models"
)

func TestDNSRecordsQtype(t *testing.T) {
	tests := []struct {
		typ   string
		qtype uint16
	}{
		{"A", dns.TypeA},
		{"AAAA", dns.TypeAAAA},
		{"MX", dns.TypeMX},
		{"CNAME", dns.TypeCNAME},
		{"TXT", dns.TypeTXT},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			rec := models.DNSRecord{Type: tt.typ}

			assert.Equal(t, tt.qtype, rec.Qtype())
		})
	}
}
