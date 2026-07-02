//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package otp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PrivacyIdeaVerifier implements the OtpVerifier interface for privacyIDEA
type PrivacyIdeaVerifier struct {
	ServerURL string
	TLSStrict bool
}

// NewPrivacyIdeaVerifier creates a new PrivacyIdeaVerifier instance
func NewPrivacyIdeaVerifier(serverURL string, tlsStrict bool) *PrivacyIdeaVerifier {
	return &PrivacyIdeaVerifier{
		ServerURL: serverURL,
		TLSStrict: tlsStrict,
	}
}

// Verify implements the OtpVerifier interface
func (p *PrivacyIdeaVerifier) Verify(userID, password, otp string) (bool, error) {
	// Example privacy idea api request:
	// curl --insecure \
	// -X POST \
	// -H "Content-Type: application/json" \
	// -d '{"user": "username", "pass": "pass+otp"}' \
	// https://idea-server.domain.lan/validate/check
	//
	// Example response for a successful authentication:
	// HTTP/1.1 200 OK
	// Content-Type: application/json
	//
	// {
	//   "detail": {
	//     "message": "matching 1 tokens",
	//     "serial": "PISP0000AB00",
	//     "type": "spass"
	//   },
	//   "id": 1,
	//   "jsonrpc": "2.0",
	//   "result": {
	//     "status": true,
	//     "value": true
	//   },
	//   "version": "privacyIDEA unknown"
	// }
	// Example response for this first part of a challenge response authentication:
	//
	// HTTP/1.1 200 OK
	// Content-Type: application/json
	//
	// {
	//   "detail": {
	// 	"serial": "PIEM0000AB00",
	// 	"type": "email",
	// 	"transaction_id": "12345678901234567890",
	// 	"multi_challenge": [ {"serial": "PIEM0000AB00",
	// 						  "transaction_id":  "12345678901234567890",
	// 						  "message": "Please enter otp from your email",
	// 						  "client_mode": "interactive"},
	// 						 {"serial": "PISM12345678",
	// 						  "transaction_id": "12345678901234567890",
	// 						  "message": "Please enter otp from your SMS",
	// 						  "client_mode": "interactive"}
	// 	]
	//   },
	//   "id": 2,
	//   "jsonrpc": "2.0",
	//   "result": {
	// 	"status": true,
	// 	"value": false
	//   },
	//   "version": "privacyIDEA unknown"
	// }

	type PrivacyIdeaResponse struct {
		Detail struct {
			Message       string `json:"message"`
			Serial        string `json:"serial"`
			Type          string `json:"type"`
			TransactionID string `json:"transaction_id"`
		} `json:"detail"`
		ID     int `json:"id"`
		Result struct {
			Status bool `json:"status"`
			Value  bool `json:"value"`
		} `json:"result"`
		Version string `json:"version"`
	}

	// Create HTTP client with TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !p.TLSStrict,
			},
		},
	}

	// Prepare request payload
	payload := map[string]string{
		"user": userID,
		"pass": password + otp,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request payload: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.ServerURL+"/validate/check", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request to privacyIDEA: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response
	var privacyResponse PrivacyIdeaResponse
	if err := json.Unmarshal(body, &privacyResponse); err != nil {
		return false, fmt.Errorf("failed to parse privacyIDEA response: %v", err)
	}

	// Check if the request was successful
	if !privacyResponse.Result.Status {
		return false, fmt.Errorf("privacyIDEA request failed: %s", privacyResponse.Detail.Message)
	}

	// Check if the OTP is valid
	if !privacyResponse.Result.Value {
		return false, fmt.Errorf("invalid OTP")
	}

	return true, nil
}
