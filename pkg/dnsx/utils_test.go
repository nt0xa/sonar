package dnsx_test

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/nt0xa/sonar/pkg/dnsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRRToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rec dns.RR
		res string
	}{
		{
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				A: net.ParseIP("127.0.0.1"),
			},
			"127.0.0.1",
		},
		{
			&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				AAAA: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			},
			"2001:db8:85a3::8a2e:370:7334",
		},
		{
			&dns.MX{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Mx:         "example.com.",
				Preference: 10,
			},
			"10 example.com.",
		},
		{
			&dns.TXT{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Txt: []string{"txt"},
			},
			"txt",
		},
		{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Target: "example.com.",
			},
			"example.com.",
		},
		{
			&dns.NS{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Ns: "ns.example.com.",
			},
			"ns.example.com.",
		},
		{
			&dns.CAA{
				Hdr: dns.RR_Header{
					Name:   "sonar.test",
					Rrtype: dns.TypeCAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Flag:  1,
				Tag:   "issue",
				Value: "letsencrypt.org",
			},
			`1 issue "letsencrypt.org"`,
		},
	}

	for _, tt := range tests {
		name := dns.Type(tt.rec.Header().Rrtype).String()

		t.Run(name, func(st *testing.T) {
			assert.Equal(t, tt.res, dnsx.RRToString(tt.rec))
		})
	}
}

func TestDNSStringToRR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		qtype uint16
		res   dns.RR
	}{
		{
			"127.0.0.1",
			dns.TypeA,
			&dns.A{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				A: net.ParseIP("127.0.0.1"),
			},
		},
		{
			"2001:db8:85a3::8a2e:370:7334",
			dns.TypeAAAA,
			&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				AAAA: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"),
			},
		},
		{
			"10 example.com.",
			dns.TypeMX,
			&dns.MX{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Mx:         "example.com.",
				Preference: 10,
			},
		},
		{
			"txt",
			dns.TypeTXT,
			&dns.TXT{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Txt: []string{"txt"},
			},
		},
		{
			"example.com.",
			dns.TypeCNAME,
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Target: "example.com.",
			},
		},
		{
			"ns.example.com.",
			dns.TypeNS,
			&dns.NS{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Ns: "ns.example.com.",
			},
		},
		{
			"0 issue \"letsencrypt.org\"",
			dns.TypeCAA,
			&dns.CAA{
				Hdr: dns.RR_Header{
					Name:   "test.sonar.test.",
					Rrtype: dns.TypeCAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(60),
				},
				Flag:  0,
				Tag:   "issue",
				Value: "letsencrypt.org",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(st *testing.T) {
			assert.Equal(
				t,
				tt.res,
				dnsx.NewRR("test.sonar.test.", tt.qtype, 60, tt.value),
			)
		})
	}
}

func TestParseRecords(t *testing.T) {
	t.Parallel()

	rrs, err := dnsx.ParseRecords(`
@ IN 60 NS ns
@ IN 60 A 127.0.0.1
@ IN 60 AAAA 2001:0db8:85a3:0000:0000:8a2e:0370:7334
@ 60 IN MX 10 mx
@ 60 IN CAA 0 issue "letsencrypt.org"
@ SOA ns1 hostmaster 1337 86400 7200 4000000 11200
`, "sonar.test")
	require.NoError(t, err)

	assert.Equal(t, rrs[0].Header().Rrtype, dns.TypeNS)
	assert.Equal(t, rrs[0].(*dns.NS).Ns, "ns.sonar.test.")

	assert.Equal(t, rrs[1].Header().Rrtype, dns.TypeA)
	assert.Equal(t, rrs[1].(*dns.A).A, net.ParseIP("127.0.0.1"))

	assert.Equal(t, rrs[2].Header().Rrtype, dns.TypeAAAA)
	assert.Equal(t, rrs[2].(*dns.AAAA).AAAA, net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334"))

	assert.Equal(t, rrs[3].Header().Rrtype, dns.TypeMX)
	assert.Equal(t, rrs[3].(*dns.MX).Mx, "mx.sonar.test.")
	assert.EqualValues(t, rrs[3].(*dns.MX).Preference, 10)

	assert.Equal(t, rrs[4].Header().Rrtype, dns.TypeCAA)
	assert.EqualValues(t, rrs[4].(*dns.CAA).Flag, 0)
	assert.Equal(t, rrs[4].(*dns.CAA).Tag, "issue")
	assert.Equal(t, rrs[4].(*dns.CAA).Value, "letsencrypt.org")

	assert.Equal(t, rrs[5].Header().Rrtype, dns.TypeSOA)
	assert.Equal(t, rrs[5].(*dns.SOA).Ns, "ns1.sonar.test.")
	assert.Equal(t, rrs[5].(*dns.SOA).Mbox, "hostmaster.sonar.test.")
}
