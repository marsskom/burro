package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func GenerateCA(certPath, keyPath string) error {
	if _, err := os.Stat(certPath); err == nil {
		return nil
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("GenerateCA: error on key generation: %w", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	tpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "Burro CA",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(
		rand.Reader,
		&tpl,
		&tpl,
		&private.PublicKey,
		private,
	)
	if err != nil {
		return fmt.Errorf("GenerateCA: error on certificate generation: %w", err)
	}

	certificate, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("GenerateCA: error to create certificate file: %w", err)
	}

	pem.Encode(certificate, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	})

	err = certificate.Close()
	if err != nil {
		return fmt.Errorf("GenerateCA: error to close certificate file: %w", err)
	}

	key, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("GenerateCA: error to create key file: %w", err)
	}

	pem.Encode(key, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})

	err = key.Close()
	if err != nil {
		return fmt.Errorf("GenerateCA: error to close key file: %w", err)
	}

	return nil
}

func LoadCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadCA: error read certificate file: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadCA: error read key file: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	keyBlock, _ := pem.Decode(keyPEM)
	if certBlock == nil || keyBlock == nil {
		return nil, nil, fmt.Errorf("LoadCA: invalid PEM format")
	}

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadCA: error parse certificate: %w", err)
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("LoadCA: error parse key: %w", err)
	}

	return caCert, caKey, nil
}

func WriteTLSCertificate(cert *tls.Certificate, certPath, keyPath string) error {
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	for _, der := range cert.Certificate {
		if err := pem.Encode(certOut, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: der,
		}); err != nil {
			return err
		}
	}

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	switch key := cert.PrivateKey.(type) {
	case *rsa.PrivateKey:
		return pem.Encode(keyOut, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})
	default:
		return fmt.Errorf("unsupported private key type %T", cert.PrivateKey)
	}
}
