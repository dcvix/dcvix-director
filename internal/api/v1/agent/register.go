// SPDX-FileCopyrightText: 2025 Diego Cortassa
// SPDX-License-Identifier: MIT

package agent

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

type registerRequest struct {
	CSR      string `json:"csr"`
	GUID     string `json:"guid"`
	Hostname string `json:"hostname"`
}

type registerResponse struct {
	Certificate            string `json:"certificate"`
	CA                     string `json:"ca"`
	AgentID                string `json:"agentId"`
	RenewalIntervalSeconds int    `json:"renewalIntervalSeconds"`
}

type registerErrorResponse struct {
	Error string `json:"error"`
}

// Register handles POST /v1/agent/register.
func Register(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	csrDER, err := base64.StdEncoding.DecodeString(req.CSR)
	if err != nil {
		http.Error(w, "invalid csr encoding", http.StatusBadRequest)
		return
	}

	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		http.Error(w, "invalid csr", http.StatusBadRequest)
		return
	}

	expectedCN := fmt.Sprintf("dcvix-agent-%s", req.GUID)
	if csr.Subject.CommonName != expectedCN {
		log.Warnf("POST /v1/agent/register: CSR CN %q does not match expected %q", csr.Subject.CommonName, expectedCN)
		http.Error(w, "csr guid mismatch", http.StatusUnauthorized)
		return
	}

	var agent models.AgentRegistration
	result := ctx.AgentDB.Where("guid = ?", req.GUID).First(&agent)

	if result.Error != nil {
		agent = models.AgentRegistration{
			GUID:     req.GUID,
			Hostname: req.Hostname,
			State:    models.AgentStatePending,
		}
		if err := ctx.AgentDB.Create(&agent).Error; err != nil {
			log.Errorf("POST /v1/agent/register: failed to create pending agent: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		log.Infof("POST /v1/agent/register: created pending agent %s (%s)", req.GUID, req.Hostname)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(registerErrorResponse{Error: "pending approval"})
		return
	}

	if agent.State == models.AgentStatePending {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(registerErrorResponse{Error: "pending approval"})
		return
	}

	if agent.State == models.AgentStateRevoked {
		http.Error(w, "agent revoked", http.StatusUnauthorized)
		return
	}

	certDER, err := ctx.Signer.SignCSR(csrDER, req.GUID, req.Hostname)
	if err != nil {
		log.Errorf("POST /v1/agent/register: failed to sign CSR: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	now := time.Now().UTC()
	ctx.AgentDB.Model(&agent).Updates(map[string]interface{}{
		"hostname":     req.Hostname,
		"last_seen_at": now,
	})

	resp := registerResponse{
		Certificate:            string(certPEM),
		CA:                     string(ctx.Signer.CACertPEM()),
		AgentID:                req.GUID,
		RenewalIntervalSeconds: 43200,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// extractAgentGUIDFromCert extracts the agent GUID from the mTLS client cert CN.
// The CN format is "dcvix-agent-<guid>".
func extractAgentGUIDFromCert(r *http.Request) (string, bool) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return "", false
	}
	cn := r.TLS.PeerCertificates[0].Subject.CommonName
	prefix := "dcvix-agent-"
	if !strings.HasPrefix(cn, prefix) {
		return "", false
	}
	return strings.TrimPrefix(cn, prefix), true
}
