//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// List handles GET /v1/sessions.
func List(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("GET /v1/sessions: request from: %s", r.RemoteAddr)

	var sessions []models.Session
	if err := ctx.DB.Order("user").Find(&sessions).Error; err != nil {
		log.Errorf("GET /v1/sessions: query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.RespondJSON(w, sessions)
}

// Create handles POST /v1/sessions (admin, JSON body).
func Create(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/sessions: request from: %s", r.RemoteAddr)

	var request struct {
		ServerID    string `json:"serverId"`
		UserID      string `json:"userId"`
		SessionID   string `json:"sessionId"`
		SessionType string `json:"sessionType"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	request.SessionID = "dcvix_" + uuid.New().String()

	url := fmt.Sprintf("https://%s:%v/v1/sessions", request.ServerID, ctx.Config.Director.AgentPort)

	requestJSON, err := json.Marshal(request)
	if err != nil {
		log.Errorf("POST /v1/sessions: failed to marshal request: %v", err)
		http.Error(w, fmt.Sprintf("Failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debugf("POST /v1/sessions: sending %s to %s", string(requestJSON), url)

	body, statusCode, err := ctx.MTLSClient.Post(url, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		log.Errorf("POST /v1/sessions: request failed: %v", err)
		http.Error(w, fmt.Sprintf("Request failed: %v", err), http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusCreated {
		log.Errorf("POST /v1/sessions: request failed with status: %v", statusCode)
		http.Error(w, fmt.Sprintf("Request failed with status: %v", statusCode), http.StatusInternalServerError)
		return
	}

	var respData map[string]string
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&respData); err != nil {
		log.Errorf("POST /v1/sessions: failed to decode response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to decode response: %v", err), http.StatusInternalServerError)
		return
	}

	api.RespondJSON(w, map[string]string{
		"sessionID": respData["sessionID"],
	})
}

// Close handles DELETE /v1/sessions/{id}.
func Close(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	log.Debugf("DELETE /v1/sessions/{id}: request for %s from: %s", sessionID, r.RemoteAddr)

	var session models.Session
	if err := ctx.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		log.Errorf("DELETE /v1/sessions/{id}: session not found: %s", sessionID)
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	url := fmt.Sprintf("https://%s:%v/v1/sessions/%s", session.ServerID, ctx.Config.Director.AgentPort, sessionID)

	log.Debugf("DELETE /v1/sessions/{id}: sending to %s", url)

	body, statusCode, err := ctx.MTLSClient.Delete(url)
	if err != nil {
		log.Errorf("DELETE /v1/sessions/{id}: failed to create request: %v", err)
		http.Error(w, fmt.Sprintf("failed to create request: %v", err), http.StatusBadGateway)
		return
	}

	if statusCode != http.StatusOK {
		uid := sessionID + "-" + session.ServerID
		if err := ctx.DB.Model(&models.Session{}).Where("uid = ?", uid).Update("status", models.StatusClosing).Error; err != nil {
			log.Errorf("DELETE /v1/sessions/{id}: failed to update session status: %v", err)
		}

		api.RespondJSON(w, map[string]interface{}{
			"status":  "error",
			"code":    statusCode,
			"message": fmt.Sprintf("server returned status: %d", statusCode),
			"body":    string(body),
		})
		return
	}

	// Set session status to closing
	uid := sessionID + "-" + session.ServerID
	if err := ctx.DB.Model(&models.Session{}).Where("uid = ?", uid).Update("status", models.StatusClosing).Error; err != nil {
		log.Errorf("DELETE /v1/sessions/{id}: failed to update session status: %v", err)
	}

	api.RespondJSON(w, map[string]interface{}{
		"status":  "success",
		"code":    statusCode,
		"message": "Session close request sent successfully",
		"body":    string(body),
	})
}
