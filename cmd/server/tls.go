package main

import (
	"crypto/tls"
	"sync"

	"github.com/bi-zone/sonar/pkg/certmanager"
)

type keypairReloader struct {
	cert *tls.Certificate

	mu sync.RWMutex
}

func (kpr *keypairReloader) reload(cert *tls.Certificate) {
	kpr.mu.Lock()
	defer kpr.mu.Unlock()
	kpr.cert = cert
}

func (kpr *keypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		kpr.mu.RLock()
		defer kpr.mu.RUnlock()
		return kpr.cert, nil
	}
}

func renewCertificate(certmgr *certmanager.CertManager, kpr *keypairReloader, domains []string) error {
	log := log.WithField("domains", domains)

	log.Infof("Trying to renew the certificate")

	newCert, err := certmgr.Renew(domains)
	if err != nil {
		return err
	}

	if newCert == nil {
		log.Infof("Certificate is not expiring")
		return nil
	}

	log.Infof("Successfully renewed the certificate")
	kpr.reload(newCert)

	return nil
}
