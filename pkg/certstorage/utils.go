package certstorage

import (
	"os"
	"strings"

	"golang.org/x/net/idna"
)

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, dirPerm)
	} else if err != nil {
		return err
	}
	return nil
}

// sanitizedDomain Make sure no funny chars are in the cert names (like wildcards ;))
func sanitizedDomain(domain string) (string, error) {
	safe, err := idna.ToASCII(strings.Replace(domain, "*", "_", -1))
	if err != nil {
		return "", err
	}
	return safe, nil
}
