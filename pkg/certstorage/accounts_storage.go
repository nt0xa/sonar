package certstorage

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"
)

const (
	baseAccountsRootFolderName = "accounts"
	baseKeysFolderName         = "keys"
	accountFileName            = "account.json"
)

// AccountsStorage A storage for account data.
//
// rootPath:
//
//     ./.lego/accounts/
//          │      └── root accounts directory
//          └── "path" option
//
// rootUserPath:
//
//     ./.lego/accounts/localhost_14000/hubert@hubert.com/
//          │      │             │             └── userID ("email" option)
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
// keysPath:
//
//     ./.lego/accounts/localhost_14000/hubert@hubert.com/keys/
//          │      │             │             │           └── root keys directory
//          │      │             │             └── userID ("email" option)
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
// accountFilePath:
//
//     ./.lego/accounts/localhost_14000/hubert@hubert.com/account.json
//          │      │             │             │             └── account file
//          │      │             │             └── userID ("email" option)
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
type AccountsStorage struct {
	caDirURL        string
	userID          string
	rootPath        string
	rootUserPath    string
	keysPath        string
	accountFilePath string
}

// NewAccountsStorage Creates a new AccountsStorage.
func NewAccountsStorage(path, email, ca string) (*AccountsStorage, error) {

	serverURL, err := url.Parse(ca)
	if err != nil {
		return nil, err
	}

	rootPath := filepath.Join(path, baseAccountsRootFolderName)
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(serverURL.Host)
	accountsPath := filepath.Join(rootPath, serverPath)
	rootUserPath := filepath.Join(accountsPath, email)

	return &AccountsStorage{
		caDirURL:        ca,
		userID:          email,
		rootPath:        rootPath,
		rootUserPath:    rootUserPath,
		keysPath:        filepath.Join(rootUserPath, baseKeysFolderName),
		accountFilePath: filepath.Join(rootUserPath, accountFileName),
	}, nil
}

func (s *AccountsStorage) ExistsAccountFilePath() bool {
	accountFile := filepath.Join(s.rootUserPath, accountFileName)
	if _, err := os.Stat(accountFile); os.IsNotExist(err) || err != nil {
		return false
	}
	return true
}

func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

func (s *AccountsStorage) GetRootUserPath() string {
	return s.rootUserPath
}

func (s *AccountsStorage) GetUserID() string {
	return s.userID
}

func (s *AccountsStorage) Save(account *Account) error {
	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.accountFilePath, jsonBytes, filePerm)
}

func (s *AccountsStorage) LoadAccount(privateKey crypto.PrivateKey) (*Account, error) {
	fileBytes, err := ioutil.ReadFile(s.accountFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not load file for account %q: %w", s.userID, err)
	}

	var account Account
	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		return nil, fmt.Errorf("could not parse file for account %q: %w", s.userID, err)
	}

	account.Key = privateKey

	if account.Registration == nil || account.Registration.Body.Status == "" {
		reg, err := tryRecoverRegistration(privateKey, s.caDirURL)
		if err != nil {
			return nil, fmt.Errorf("could not load account for %q. registration is nil: %w", s.userID, err)
		}

		account.Registration = reg
		err = s.Save(&account)
		if err != nil {
			return nil, fmt.Errorf("could not save account for %q. registration is nil: %w", s.userID, err)
		}
	}

	return &account, nil
}

func (s *AccountsStorage) GetPrivateKey(keyType certcrypto.KeyType) (crypto.PrivateKey, error) {
	accKeyPath := filepath.Join(s.keysPath, s.userID+".key")

	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		if err := s.createKeysFolder(); err != nil {
			return nil, fmt.Errorf("could not create directory for keys: %w", err)
		}

		privateKey, err := generatePrivateKey(accKeyPath, keyType)
		if err != nil {
			return nil, fmt.Errorf("could not generate rsa private account key for account %q: %w", s.userID, err)
		}

		return privateKey, nil
	}

	privateKey, err := loadPrivateKey(accKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not load rsa private key from file %q: %w", accKeyPath, err)
	}

	return privateKey, nil
}

func (s *AccountsStorage) createKeysFolder() error {
	if err := createNonExistingFolder(s.keysPath); err != nil {
		return err
	}
	return nil
}

func generatePrivateKey(file string, keyType certcrypto.KeyType) (crypto.PrivateKey, error) {
	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, err
	}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	defer certOut.Close()

	pemKey := certcrypto.PEMBlock(privateKey)
	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private key type")
}

func tryRecoverRegistration(privateKey crypto.PrivateKey, CADirURL string) (*registration.Resource, error) {
	// couldn't load account but got a key. Try to look the account up.
	config := lego.NewConfig(&Account{Key: privateKey})

	config.CADirURL = CADirURL

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	reg, err := client.Registration.ResolveAccountByKey()
	if err != nil {
		return nil, err
	}
	return reg, nil
}
