package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

func GenerateHostCertificate(
	host string,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) (*tls.Certificate, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	tpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: host,
		},
		DNSNames:  []string{host},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	der, err := x509.CreateCertificate(
		rand.Reader,
		&tpl,
		caCert,
		&private.PublicKey,
		caKey,
	)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(private),
	})

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &certificate, err
}
