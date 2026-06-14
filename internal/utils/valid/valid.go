package valid

import (
	"errors"
	"os"
)

func File(value any) error {
	path, _ := value.(string)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func Directory(value any) error {
	path, _ := value.(string)

	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if fi.Mode().IsRegular() {
		return errors.New("must be directory")
	}

	return nil
}
