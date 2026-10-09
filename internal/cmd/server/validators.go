package server

import (
	"errors"
	"net/netip"
	"os"
)

// file asserts the path exists on disk.
func file(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// directory asserts the path is not a regular file (i.e. a directory or absent).
func directory(path string) error {
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if fi.Mode().IsRegular() {
		return errors.New("must be directory")
	}
	return nil
}

// prefix asserts the string is a CIDR prefix like 10.0.0.0/8.
func prefix(s string) error {
	if _, err := netip.ParsePrefix(s); err != nil {
		return errors.New("must be a valid CIDR prefix")
	}
	return nil
}
