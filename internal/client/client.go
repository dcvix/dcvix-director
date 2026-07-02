//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type MTLSClient struct {
	httpClient *http.Client
}

func NewMTLSClient(certFile, keyFile, caFile string) (*MTLSClient, error) {
	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("error loading client key pair: %w", err)
	}

	// Load the CA certificate
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Create custom verification function to avoid hostname/IP verification
	verifyPeerCert := func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		opts := x509.VerifyOptions{
			Roots:         caCertPool,
			CurrentTime:   time.Now(),
			Intermediates: x509.NewCertPool(),
		}
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %v", err)
		}
		_, err = cert.Verify(opts)
		return err
	}

	tlsConfig := &tls.Config{
		Certificates:          []tls.Certificate{clientCert},
		RootCAs:               caCertPool,
		MinVersion:            tls.VersionTLS12,
		InsecureSkipVerify:    false,
		VerifyPeerCertificate: verifyPeerCert,
	}

	// Create an HTTP client with the custom TLS configuration
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 10 * time.Second,
	}

	return &MTLSClient{httpClient: httpClient}, nil
}

func (c *MTLSClient) Get(url string) ([]byte, int, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (c *MTLSClient) Post(url string, contentType string, body io.Reader) ([]byte, int, error) {
	resp, err := c.httpClient.Post(url, contentType, body)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func (c *MTLSClient) Delete(url string) ([]byte, int, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}
