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
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

type renewRequest struct {
	CSR string `json:"csr"`
}

type renewResponse struct {
	Certificate            string `json:"certificate"`
	RenewalIntervalSeconds int    `json:"renewalIntervalSeconds"`
}

// Renew handles POST /v1/agent/renew (mTLS).
func Renew(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	agentGUID, ok := extractAgentGUIDFromCert(r)
	if !ok {
		http.Error(w, "invalid client certificate", http.StatusUnauthorized)
		return
	}

	var req renewRequest
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

	expectedCN := fmt.Sprintf("dcvix-agent-%s", agentGUID)
	if csr.Subject.CommonName != expectedCN {
		log.Warnf("POST /v1/agent/renew: CSR CN %q does not match authenticated GUID %q", csr.Subject.CommonName, agentGUID)
		http.Error(w, "csr guid mismatch", http.StatusUnauthorized)
		return
	}

	var agent models.AgentRegistration
	result := ctx.AgentDB.Where("guid = ?", agentGUID).First(&agent)
	if result.Error != nil {
		log.Warnf("POST /v1/agent/renew: unknown agent %s", agentGUID)
		http.Error(w, "unknown agent", http.StatusUnauthorized)
		return
	}

	if agent.State == models.AgentStateRevoked {
		log.Warnf("POST /v1/agent/renew: revoked agent %s attempted renewal", agentGUID)
		http.Error(w, "agent revoked", http.StatusUnauthorized)
		return
	}

	if agent.State != models.AgentStateRegistered {
		agent.State = models.AgentStateRegistered
		now := time.Now().UTC()
		ctx.AgentDB.Model(&agent).Updates(map[string]interface{}{
			"state":         models.AgentStateRegistered,
			"registered_at": now,
		})
	}

	certDER, err := ctx.Signer.SignCSR(csrDER, agentGUID, agent.Hostname)
	if err != nil {
		log.Errorf("POST /v1/agent/renew: failed to sign CSR: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	now := time.Now().UTC()
	ctx.AgentDB.Model(&agent).Updates(map[string]interface{}{
		"last_seen_at": now,
	})

	resp := renewResponse{
		Certificate:            string(certPEM),
		RenewalIntervalSeconds: 43200,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
