// SPDX-FileCopyrightText: 2025 Diego Cortassa
// SPDX-License-Identifier: MIT

package ca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Signer struct {
	caCert    *x509.Certificate
	caCertPEM []byte
	caKey     crypto.PrivateKey
	dataDir   string
}

// NewSigner loads the CA cert+key from {dataDir}/ca.pem and {dataDir}/ca.key.
// If they don't exist, GenerateCA is called first.
func NewSigner(dataDir string) (*Signer, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	caKeyPath := filepath.Join(dataDir, "ca.key")
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		if err := GenerateCA(dataDir); err != nil {
			return nil, fmt.Errorf("failed to generate CA: %w", err)
		}
	}

	if err := checkKeyPermissions(caKeyPath); err != nil {
		return nil, err
	}

	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca.key: %w", err)
	}
	block, _ := pem.Decode(caKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode ca.key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca.key: %w", err)
	}

	caCertPEM, err := os.ReadFile(filepath.Join(dataDir, "ca.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ca.pem: %w", err)
	}
	block, _ = pem.Decode(caCertPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode ca.pem PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ca.pem: %w", err)
	}

	serverCertPath := filepath.Join(dataDir, "server.crt")
	serverKeyPath := filepath.Join(dataDir, "server.key")
	if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
		log.Info("Server cert not found, generating new server cert")
		if err := generateServerCert(dataDir, cert, key); err != nil {
			return nil, fmt.Errorf("failed to generate server cert: %w", err)
		}
	} else if _, err := os.Stat(serverKeyPath); os.IsNotExist(err) {
		log.Warn("Server key missing, regenerating server cert")
		if err := os.Remove(serverCertPath); err != nil {
			return nil, fmt.Errorf("failed to remove stale server.crt: %w", err)
		}
		if err := generateServerCert(dataDir, cert, key); err != nil {
			return nil, fmt.Errorf("failed to generate server cert: %w", err)
		}
	}

	return &Signer{caCert: cert, caCertPEM: caCertPEM, caKey: key, dataDir: dataDir}, nil
}

// GenerateCA creates a new ECDSA P-256 CA key pair and self-signed CA certificate.
// Writes ca.pem (0644) and ca.key (0600) to dataDir.
// Also generates the server cert+key (server.crt, server.key) signed by this CA.
func GenerateCA(dataDir string) error {
	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return err
	}

	now := time.Now()
	caTmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "dcvix CA"},
		NotBefore:             now.Add(-1 * time.Hour),
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caPriv.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caKeyPEM := &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes(caPriv)}
	caCertPEM := &pem.Block{Type: "CERTIFICATE", Bytes: caDER}

	if err := os.WriteFile(filepath.Join(dataDir, "ca.key"), pem.EncodeToMemory(caKeyPEM), 0600); err != nil {
		return fmt.Errorf("failed to write ca.key: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "ca.pem"), pem.EncodeToMemory(caCertPEM), 0644); err != nil {
		return fmt.Errorf("failed to write ca.pem: %w", err)
	}

	return generateServerCert(dataDir, caTmpl, caPriv)
}

func serverCertHostnames() []string {
	names := []string{"dcvix-director", "localhost"}
	if h, err := os.Hostname(); err == nil && h != "" {
		names = append(names, h)
		if fqdn, err := net.LookupCNAME(h); err == nil {
			fqdn = strings.TrimSuffix(fqdn, ".")
			if fqdn != h && fqdn != "" {
				names = append(names, fqdn)
			}
		}
	}
	return names
}

func generateServerCert(dataDir string, caTmpl *x509.Certificate, caPriv crypto.PrivateKey) error {
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate server key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return err
	}

	now := time.Now()
	serverTmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "dcvix-director"},
		DNSNames:     serverCertHostnames(),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    now.Add(-1 * time.Hour),
		NotAfter:     now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caPriv)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %w", err)
	}

	serverKeyPEM := &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes(serverKey)}
	serverCertPEM := &pem.Block{Type: "CERTIFICATE", Bytes: serverDER}

	if err := os.WriteFile(filepath.Join(dataDir, "server.key"), pem.EncodeToMemory(serverKeyPEM), 0600); err != nil {
		return fmt.Errorf("failed to write server.key: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "server.crt"), pem.EncodeToMemory(serverCertPEM), 0644); err != nil {
		return fmt.Errorf("failed to write server.crt: %w", err)
	}

	return nil
}

// SignCSR takes a DER-encoded CSR and produces a DER-encoded signed certificate
// with 14-day validity, KeyUsage digitalSignature, and ExtKeyUsage serverAuth+clientAuth.
func (s *Signer) SignCSR(csrDER []byte, agentID string, hostname string) ([]byte, error) {
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return nil, fmt.Errorf("invalid CSR: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return nil, err
	}

	dnsNames := make([]string, 0, len(csr.DNSNames)+1)
	seen := map[string]bool{}
	for _, n := range csr.DNSNames {
		if !seen[n] {
			dnsNames = append(dnsNames, n)
			seen[n] = true
		}
	}
	if hostname != "" && !seen[hostname] {
		dnsNames = append(dnsNames, hostname)
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      csr.Subject,
		DNSNames:     dnsNames,
		NotBefore:    now.Add(-1 * time.Hour),
		NotAfter:     now.Add(14 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	return x509.CreateCertificate(rand.Reader, tmpl, s.caCert, csr.PublicKey, s.caKey)
}

// SignServerCert generates a server certificate signed by the CA for the given public key.
func (s *Signer) SignServerCert(publicKey crypto.PublicKey) ([]byte, error) {
	serial, err := randomSerial()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "dcvix-director"},
		DNSNames:     serverCertHostnames(),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    now.Add(-1 * time.Hour),
		NotAfter:     now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	return x509.CreateCertificate(rand.Reader, tmpl, s.caCert, publicKey, s.caKey)
}

// CACertPool returns the CA certificate as an x509.CertPool.
func (s *Signer) CACertPool() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(s.caCert)
	return pool
}

// CACertificate returns the parsed CA certificate.
func (s *Signer) CACertificate() *x509.Certificate {
	return s.caCert
}

// CACertPEM returns the PEM-encoded CA certificate bytes.
func (s *Signer) CACertPEM() []byte {
	return s.caCertPEM
}

func checkKeyPermissions(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&0o077 != 0 {
		return fmt.Errorf("key file %s has group/other permissions (mode %o); must be 0600", path, fi.Mode())
	}
	return nil
}

func randomSerial() (*big.Int, error) {
	serial := make([]byte, 16)
	if _, err := rand.Read(serial); err != nil {
		return nil, fmt.Errorf("failed to generate serial: %w", err)
	}
	serial[0] &= 0x7f
	return new(big.Int).SetBytes(serial), nil
}

func privBytes(key crypto.PrivateKey) []byte {
	bytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}
	return bytes
}
