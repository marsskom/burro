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
	"net"
	"time"
)

func GenerateHostCertificate(
	host string,
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
) (*tls.Certificate, error) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("host cert: error generate key: %w", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	DNSNames := []string{hostname}
	if hostname == "localhost" {
		DNSNames = []string{
			hostname,
			"127.0.0.1",
			"::1",
		}
	}

	if port != "" {
		DNSNames = append(DNSNames, fmt.Sprintf(":%s", port))
	}

	tpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: hostname,
		},
		DNSNames:  DNSNames,
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(1, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
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
		return nil, fmt.Errorf("host cert: error create certificate: %w", err)
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
		return nil, fmt.Errorf("host cert: error parse x509 key pair: %w", err)
	}

	return &certificate, nil
}
