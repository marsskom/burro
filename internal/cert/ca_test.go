package cert

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateCA_CreatesFiles(t *testing.T) {
	dir := t.TempDir()

	certPath := filepath.Join(dir, "ca.pem")
	keyPath := filepath.Join(dir, "ca.key")

	err := GenerateCA(certPath, keyPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Fatalf("certificate file not created: %v", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key file not created: %v", err)
	}
}

func TestGenerateCA_Idempotent(t *testing.T) {
	dir := t.TempDir()

	certPath := filepath.Join(dir, "ca.pem")
	keyPath := filepath.Join(dir, "ca.key")

	if err := GenerateCA(certPath, keyPath); err != nil {
		t.Fatalf("first generate failed: %v", err)
	}

	// Gets files data.
	c1Stat, _ := os.Stat(certPath)
	c1, _ := os.ReadFile(certPath)

	k1Stat, _ := os.Stat(keyPath)
	k1, _ := os.ReadFile(keyPath)

	if err := GenerateCA(certPath, keyPath); err != nil {
		t.Fatalf("second generate failed: %v", err)
	}

	c2Stat, _ := os.Stat(certPath)
	c2, _ := os.ReadFile(certPath)

	k2Stat, _ := os.Stat(keyPath)
	k2, _ := os.ReadFile(keyPath)

	if c1Stat.ModTime() != c2Stat.ModTime() {
		t.Fatalf("certificate file has been modified")
	}

	if string(c1) != string(c2) {
		t.Fatal("certificate was unexpectedly overwritten")
	}

	if k1Stat.ModTime() != k2Stat.ModTime() {
		t.Fatal("key fiel has been modified")
	}

	if string(k1) != string(k2) {
		t.Fatal("key was unexpectedly overwritten")
	}
}

func TestLoadCA_Success(t *testing.T) {
	dir := t.TempDir()

	certPath := filepath.Join(dir, "ca.pem")
	keyPath := filepath.Join(dir, "ca.key")

	if err := GenerateCA(certPath, keyPath); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	cert, key, err := LoadCA(certPath, keyPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cert == nil || key == nil {
		t.Fatal("expected non-nil cert and key")
	}

	if cert.Subject.CommonName != "Burro CA" {
		t.Fatalf("unexpected CN: %s", cert.Subject.CommonName)
	}

	if !cert.IsCA {
		t.Fatal("certificate is not CA")
	}
}

func TestLoadCA_InvalidCertPath(t *testing.T) {
	_, _, err := LoadCA("/non/existing/cert.pem", "/non/existing/key.pem")
	if err == nil {
		t.Fatal("expected error for missing files")
	}
}

func TestLoadCA_InvalidPEM(t *testing.T) {
	dir := t.TempDir()

	certPath := filepath.Join(dir, "bad.pem")
	keyPath := filepath.Join(dir, "bad.key")

	_ = os.WriteFile(certPath, []byte("invalid cert"), 0644)
	_ = os.WriteFile(keyPath, []byte("invalid key"), 0644)

	_, _, err := LoadCA(certPath, keyPath)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestGenerateCA_ValidCertificateParsing(t *testing.T) {
	dir := t.TempDir()

	certPath := filepath.Join(dir, "ca.pem")
	keyPath := filepath.Join(dir, "ca.key")

	if err := GenerateCA(certPath, keyPath); err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	// Decodes and parses directly.
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode PEM")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("invalid certificate generated: %v", err)
	}
}
