package storage

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/go-acme/lego/v3/certcrypto"
)

func pemEncode(data interface{}) ([]byte, error) {
	var pemBytes bytes.Buffer

	pemKey := certcrypto.PEMBlock(data)
	if pemKey == nil {
		return nil, fmt.Errorf("invalid data")
	}

	if err := pem.Encode(&pemBytes, pemKey); err != nil {
		return nil, fmt.Errorf("fail to encode data key: %w", err)
	}

	return pemBytes.Bytes(), nil

}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
