//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

// ResolveSession handles POST /resolveSession.
func ResolveSession(ctx *HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /resolveSession: request from: %s", r.RemoteAddr)

	var request struct {
		SessionID         string `json:"sessionId"`
		TransportProtocol string `json:"transport"`
		ClientIPAddress   string `json:"clientIpAddress,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Errorf("POST /resolveSession: invalid request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	session := models.Session{}
	if err := ctx.DB.First(&session, "id = ?", request.SessionID).Error; err != nil {
		log.Debugf("POST /resolveSession: session not found: %s", request.SessionID)
		http.Error(w, "ResolveSession: session not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, map[string]string{
		"sessionId":         request.SessionID,
		"transportProtocol": request.TransportProtocol,
		"dcvServerEndpoint": session.ServerID,
		"port":              "8443", // TODO: Where do I get the port?
		"webUrlPath":        "/",
	})
}
