package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
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
		return err
	}

	tpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Burro CA",
		},
		NotBefore:             time.Now(),
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
		return err
	}

	certificate, err := os.Create(certPath)
	if err != nil {
		return err
	}

	pem.Encode(certificate, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	})

	err = certificate.Close()
	if err != nil {
		return err
	}

	key, err := os.Create(keyPath)
	if err != nil {
		return err
	}

	pem.Encode(key, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})

	return key.Close()
}

func LoadCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, err
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}

	certBlock, _ := pem.Decode(certPEM)
	keyBlock, _ := pem.Decode(keyPEM)

	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	caKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}
