package geoipx_test

import (
	"context"
	"io"
	"log/slog"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nt0xa/sonar/pkg/geoipx"
)

func Test_Geox(t *testing.T) {
	log := slog.New(slog.DiscardHandler)

	// Create temporary copies of test files
	tempDir := t.TempDir()
	cityPath := filepath.Join(tempDir, "city.mmdb")
	asnPath := filepath.Join(tempDir, "asn.mmdb")

	// Copy empty test files initially
	copyFile(t, "test-data/GeoLite2-City-Test-Empty.mmdb", cityPath)
	copyFile(t, "test-data/GeoLite2-ASN-Test-Empty.mmdb", asnPath)

	// Create DB with temporary files
	db, err := geoipx.New(log, cityPath, asnPath)
	require.NoError(t, err)
	defer func() {
		_ = db.Close()
	}()

	// Start watching files
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.Watch(ctx)
	require.NoError(t, err)

	// Initial lookup with empty database - should return empty results
	ip := netip.MustParseAddr("81.2.69.142")
	info, err := db.Lookup(ip)
	require.NoError(t, err)
	assert.Empty(t, info.City) // Empty database has no city data

	// Replace files with the full test databases
	copyFile(t, "test-data/GeoLite2-City-Test.mmdb", cityPath)
	copyFile(t, "test-data/GeoLite2-ASN-Test.mmdb", asnPath)

	// Wait a bit for file watcher to detect changes and reload
	time.Sleep(500 * time.Millisecond)

	// Now lookup should return data from the new database files
	info, err = db.Lookup(ip)
	require.NoError(t, err)
	assert.Equal(t, "London", info.City)
	assert.Equal(t, "United Kingdom", info.Country.Name)
	assert.Equal(t, "GB", info.Country.ISOCode)
	assert.ElementsMatch(t, []string{"England"}, info.Subdivisions)

	// Test ASN lookup
	asnIP := netip.MustParseAddr("1.0.0.1")
	info, err = db.Lookup(asnIP)
	require.NoError(t, err)
	assert.EqualValues(t, 15169, info.ASN.Number)
	assert.Equal(t, "Google Inc.", info.ASN.Org)
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	srcFile, err := os.Open(src)
	require.NoError(t, err)
	defer func() {
		_ = srcFile.Close()
	}()

	dstFile, err := os.Create(dst)
	require.NoError(t, err)
	defer func() {
		_ = dstFile.Close()
	}()

	_, err = io.Copy(dstFile, srcFile)
	require.NoError(t, err)
}
