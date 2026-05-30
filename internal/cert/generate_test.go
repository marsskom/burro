package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

func TestGenerateHostCertificate_BasicValidity(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	host := "example.com"

	cert, err := GenerateHostCertificate(host, caCert, caKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cert == nil {
		t.Fatal("expected certificate, got nil")
	}
}

func TestGenerateHostCertificate_X509Parsing(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	cert, err := GenerateHostCertificate("example.com", caCert, caKey)
	if err != nil {
		t.Fatal(err)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("invalid x509 cert: %v", err)
	}

	if x509Cert.Subject.CommonName != "example.com" {
		t.Fatalf("expected CN example.com, got %s", x509Cert.Subject.CommonName)
	}
}

func TestGenerateHostCertificate_DNSNames(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	cert, err := GenerateHostCertificate("example.com", caCert, caKey)
	if err != nil {
		t.Fatal(err)
	}

	x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])

	if len(x509Cert.DNSNames) != 1 || x509Cert.DNSNames[0] != "example.com" {
		t.Fatalf("invalid DNSNames: %+v", x509Cert.DNSNames)
	}
}

func TestGenerateHostCertificate_SignedByCA(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	cert, err := GenerateHostCertificate("example.com", caCert, caKey)
	if err != nil {
		t.Fatal(err)
	}

	x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])

	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	if _, err := x509Cert.Verify(opts); err != nil {
		t.Fatalf("certificate not signed by CA: %v", err)
	}
}

func TestGenerateHostCertificate_ValidityPeriod(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	cert, err := GenerateHostCertificate("example.com", caCert, caKey)
	if err != nil {
		t.Fatal(err)
	}

	x509Cert, _ := x509.ParseCertificate(cert.Certificate[0])

	now := time.Now()

	if x509Cert.NotBefore.After(now) {
		t.Fatal("NotBefore is in future")
	}

	if x509Cert.NotAfter.Before(now.AddDate(0, 11, 0)) {
		t.Fatal("certificate validity too short")
	}
}

func TestGenerateHostCertificate_UniqueSerials(t *testing.T) {
	caCert, caKey := generateTestCA(t)

	c1, _ := GenerateHostCertificate("a.com", caCert, caKey)
	c2, _ := GenerateHostCertificate("b.com", caCert, caKey)

	x1, _ := x509.ParseCertificate(c1.Certificate[0])
	x2, _ := x509.ParseCertificate(c2.Certificate[0])

	if x1.SerialNumber.Cmp(x2.SerialNumber) == 0 {
		t.Fatal("expected unique serial numbers")
	}
}

func generateTestCA(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	caCert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	return caCert, priv
}
