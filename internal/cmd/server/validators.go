package server

import (
	"errors"
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
