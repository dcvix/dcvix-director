//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package servers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	log "github.com/sirupsen/logrus"
)

// ConfigEntry represents a single configuration parameter.
type ConfigEntry struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// SetConfig handles POST /v1/user/servers/{server}/config.
func SetConfig(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	serverID := r.PathValue("server")
	log.Debugf("POST /v1/user/servers/{server}/config: received for server %s from: %s", serverID, r.RemoteAddr)

	var request struct {
		Config []ConfigEntry `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Errorf("POST /v1/user/servers/{server}/config: invalid request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	forwardPayload := struct {
		Config []ConfigEntry `json:"config"`
	}{Config: request.Config}

	forwardJSON, err := json.Marshal(forwardPayload)
	if err != nil {
		log.Errorf("POST /v1/user/servers/{server}/config: failed to marshal request: %v", err)
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("https://%s:%v/v1/config", serverID, ctx.Config.Director.AgentPort)

	log.Debugf("POST /v1/user/servers/{server}/config: sending %s to %s", string(forwardJSON), url)

	_, statusCode, err := ctx.MTLSClient.Post(url, "application/json", bytes.NewBuffer(forwardJSON))
	if err != nil {
		log.Errorf("POST /v1/user/servers/{server}/config: request failed: %v", err)
		http.Error(w, fmt.Sprintf("Request failed: %v", err), http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusOK {
		log.Errorf("POST /v1/user/servers/{server}/config: agent returned status: %v", statusCode)
		http.Error(w, fmt.Sprintf("Agent returned status: %v", statusCode), statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
}
