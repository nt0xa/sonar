package certstorage

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/certificate"
)

const (
	baseCertificatesFolderName = "certificates"
	baseArchivesFolderName     = "archives"
)

const (
	filePerm os.FileMode = 0600
	dirPerm  os.FileMode = 0700
)

// CertificatesStorage a certificates storage.
//
// rootPath:
//
//     ./.lego/certificates/
//          │      └── root certificates directory
//          └── "path" option
//
// archivePath:
//
//     ./.lego/archives/
//          │      └── archived certificates directory
//          └── "path" option
//
type CertificatesStorage struct {
	rootPath    string
	archivePath string
	pem         bool
}

// NewCertificatesStorage create a new certificates storage.
func NewCertificatesStorage(path string, pem bool) *CertificatesStorage {
	return &CertificatesStorage{
		rootPath:    filepath.Join(path, baseCertificatesFolderName),
		archivePath: filepath.Join(path, baseArchivesFolderName),
		pem:         pem,
	}
}

func (s *CertificatesStorage) CreateRootFolder() error {
	err := createNonExistingFolder(s.rootPath)
	if err != nil {
		return fmt.Errorf("could not check/create path: %w", err)
	}

	return nil
}

func (s *CertificatesStorage) CreateArchiveFolder() error {
	err := createNonExistingFolder(s.archivePath)
	if err != nil {
		return fmt.Errorf("could not check/create path: %w", err)
	}
	return nil
}

func (s *CertificatesStorage) GetRootPath() string {
	return s.rootPath
}

func (s *CertificatesStorage) SaveResource(certRes *certificate.Resource) error {
	domain := certRes.Domain

	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	err := s.WriteFile(domain, ".crt", certRes.Certificate)
	if err != nil {
		return fmt.Errorf("unable to save certificate for domain %q: %w", domain, err)
	}

	if certRes.IssuerCertificate != nil {
		err = s.WriteFile(domain, ".issuer.crt", certRes.IssuerCertificate)
		if err != nil {
			return fmt.Errorf("unable to save issuer certificate for domain %q: %w", domain, err)
		}
	}

	if certRes.PrivateKey != nil {
		// if we were given a CSR, we don't know the private key
		err = s.WriteFile(domain, ".key", certRes.PrivateKey)
		if err != nil {
			return fmt.Errorf("unable to save private key for domain %q: %w", domain, err)
		}

		if s.pem {
			err = s.WriteFile(domain, ".pem", bytes.Join([][]byte{certRes.Certificate, certRes.PrivateKey}, nil))
			if err != nil {
				return fmt.Errorf("unable to save certificate and private key in .pem for domain %q: %w", domain, err)
			}
		}
	} else if s.pem {
		// we don't have the private key; can't write the .pem file
		return fmt.Errorf("unable to save pem withou private key for domain %q: %w", domain, err)
	}

	jsonBytes, err := json.MarshalIndent(certRes, "", "\t")
	if err != nil {
		return fmt.Errorf("unable to marshal cert resource for domain %q: %w", domain, err)
	}

	err = s.WriteFile(domain, ".json", jsonBytes)
	if err != nil {
		return fmt.Errorf("unable to save cert resource for domain %q: %w", domain, err)
	}

	return nil
}

func (s *CertificatesStorage) ReadResource(domain string) (*certificate.Resource, error) {
	raw, err := s.ReadFile(domain, ".json")
	if err != nil {
		return nil, fmt.Errorf("error while loading the meta data for domain %q: %w", domain, err)
	}

	var resource certificate.Resource
	if err = json.Unmarshal(raw, &resource); err != nil {
		return nil, fmt.Errorf("error while marshaling the meta data for domain %q: %w", domain, err)
	}

	return &resource, nil
}

func (s *CertificatesStorage) ExistsFile(domain, extension string) bool {
	filename, err := sanitizedDomain(domain)
	if err != nil {
		return false
	}
	filename += extension
	filePath := filepath.Join(s.rootPath, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) || err != nil {
		return false
	}

	return true
}

func (s *CertificatesStorage) ReadFile(domain, extension string) ([]byte, error) {
	filename, err := sanitizedDomain(domain)
	if err != nil {
		return nil, err
	}
	filename += extension
	filePath := filepath.Join(s.rootPath, filename)

	return ioutil.ReadFile(filePath)
}

func (s *CertificatesStorage) ReadCertificate(domain, extension string) ([]*x509.Certificate, error) {
	content, err := s.ReadFile(domain, extension)
	if err != nil {
		return nil, err
	}

	// The input may be a bundle or a single certificate.
	return certcrypto.ParsePEMBundle(content)
}

func (s *CertificatesStorage) WriteFile(domain, extension string, data []byte) error {
	baseFileName, err := sanitizedDomain(domain)
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.rootPath, baseFileName+extension)

	return ioutil.WriteFile(filePath, data, filePerm)
}

func (s *CertificatesStorage) MoveToArchive(domain string) error {
	filename, err := sanitizedDomain(domain)
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(s.rootPath, filename+".*"))
	if err != nil {
		return err
	}

	for _, oldFile := range matches {
		date := strconv.FormatInt(time.Now().Unix(), 10)
		filename := date + "." + filepath.Base(oldFile)
		newFile := filepath.Join(s.archivePath, filename)

		err = os.Rename(oldFile, newFile)
		if err != nil {
			return err
		}
	}

	return nil
}
