//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Create handles POST /v1/user/sessions (user, JSON body).
func Create(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/user/sessions: request from: %s", r.RemoteAddr)

	userID, ok := api.UserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		ServerID    string `json:"serverId"`
		SessionType string `json:"sessionType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Errorf("POST /v1/user/sessions: invalid request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// TODO: if GetPoolsForUser() in internal/database/policy.go is changed returns pools instead of the workstations in the pools, we could open a session on a round robin or random or free server

	// Check if user is authorized to connect to server
	authorized := ctx.PolicyDB.IsAdmin(userID)
	if !authorized {
		for _, s := range ctx.PolicyDB.GetServers(userID) {
			if s == "ALLOW_CUSTOM" || strings.EqualFold(s, request.ServerID) {
				authorized = true
				break
			}
		}
	}
	if !authorized {
		log.Warnf("POST /v1/user/sessions: user %s not authorized for server %s", userID, request.ServerID)
		http.Error(w, "Forbidden: not authorized for this server", http.StatusForbidden)
		return
	}

	// TODO: check if sessiontype is valid

	// Check if a session already exists for the user on the requested server and of the requested type
	verifyURL := fmt.Sprintf("https://%s:%v/v1/sessions", request.ServerID, ctx.Config.Director.AgentPort)
	log.Debugf("POST /v1/user/sessions: listing sessions at %s", verifyURL)
	body, statusCode, err := ctx.MTLSClient.Get(verifyURL)
	if err != nil {
		log.Errorf("POST /v1/user/sessions: failed to get sessions list: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get sessions list: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("POST /v1/user/sessions: status code: %d", statusCode)
	if statusCode == http.StatusOK {
		var existing []models.Session
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&existing); err != nil {
			log.Errorf("POST /v1/user/sessions: failed to decode sessions: %v", err)
			http.Error(w, fmt.Sprintf("Failed to decode sessions: %v", err), http.StatusInternalServerError)
			return
		}
		log.Debugf("POST /v1/user/sessions: sessions list: %v", existing)
		for _, s := range existing {
			log.Debugf("POST /v1/user/sessions: checking session %s, owner: %s, type: %s vs user: %s, type: %s",
				s.ID, s.Owner, s.Type, userID, request.SessionType)
			if s.Owner == userID && strings.EqualFold(s.Type, request.SessionType) {
				log.Debugf("POST /v1/user/sessions: reusing existing session %s for user %s on server %s",
					s.ID, userID, request.ServerID)
				api.RespondJSON(w, map[string]string{
					"sessionID": s.ID,
				})
				return
			}
		}
	}

	// create a new session
	sessionID := "dcvix_" + uuid.New().String()

	forwardReq := map[string]string{
		"serverId":    request.ServerID,
		"userId":      userID,
		"sessionId":   sessionID,
		"sessionType": request.SessionType,
	}

	forwardJSON, err := json.Marshal(forwardReq)
	if err != nil {
		log.Errorf("POST /v1/user/sessions: failed to marshal request: %v", err)
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("https://%s:%v/v1/sessions", request.ServerID, ctx.Config.Director.AgentPort)

	log.Debugf("POST /v1/user/sessions: sending %s to %s", string(forwardJSON), url)

	// agent answer with session id
	body, statusCode, err = ctx.MTLSClient.Post(url, "application/json", bytes.NewBuffer(forwardJSON))
	if err != nil {
		log.Errorf("POST /v1/user/sessions: request failed: %v", err)
		http.Error(w, fmt.Sprintf("Request failed: %v", err), http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusCreated {
		log.Errorf("POST /v1/user/sessions: request failed with status: %v", statusCode)
		http.Error(w, fmt.Sprintf("Request failed with status: %v", statusCode), http.StatusInternalServerError)
		return
	}

	var respData map[string]string
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&respData); err != nil {
		log.Errorf("POST /v1/user/sessions: failed to decode response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	sessionID = respData["sessionID"]
	api.RespondJSON(w, map[string]string{
		"sessionID": sessionID,
	})
}
