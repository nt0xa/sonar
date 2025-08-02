package geox_test

import (
	"net/netip"
	"testing"

	"github.com/nt0xa/sonar/pkg/geox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Geox(t *testing.T) {
	gdb, err := geox.New(
		"test-data/GeoLite2-City-Test.mmdb",
		"test-data/GeoLite2-ASN-Test.mmdb",
	)
	require.NoError(t, err)

	info, err := gdb.Lookup(netip.MustParseAddr("81.2.69.142"))
	require.NoError(t, err)

	assert.Equal(t, "London", info.City)
	assert.Equal(t, "United Kingdom", info.Country.Name)
	assert.Equal(t, "GB", info.Country.ISOCode)
	assert.Equal(t, "🇬🇧", info.Country.Flag)
	assert.Equal(t, "England", info.Subdivisions[0])

	info, err = gdb.Lookup(netip.MustParseAddr("1.0.0.1"))
	require.NoError(t, err)

	assert.EqualValues(t, 15169, info.ASN.Number)
	assert.Equal(t, "Google Inc.", info.ASN.Org)
}
