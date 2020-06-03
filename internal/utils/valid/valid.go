package valid

import (
	"errors"
	"os"
)

func File(value interface{}) error {
	path, _ := value.(string)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func Directory(value interface{}) error {
	path, _ := value.(string)

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return err
	} else if fi.Mode().IsRegular() {
		return errors.New("must be directory")
	}

	return nil
}
