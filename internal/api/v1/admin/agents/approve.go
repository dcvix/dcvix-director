// SPDX-FileCopyrightText: 2025 Diego Cortassa
// SPDX-License-Identifier: MIT

package agents

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

// ListPending handles GET /v1/admin/agents (lists all agents).
func ListPending(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	stateFilter := r.URL.Query().Get("state")

	var agents []models.AgentRegistration
	query := ctx.AgentDB.Order("created_at desc")
	if stateFilter != "" {
		query = query.Where("state = ?", stateFilter)
	}
	if err := query.Find(&agents).Error; err != nil {
		log.Errorf("GET /v1/admin/agents: query failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// Approve handles POST /v1/admin/agents/{guid}/approve.
func Approve(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	guid := r.PathValue("guid")
	if guid == "" {
		http.Error(w, "guid required", http.StatusBadRequest)
		return
	}

	result := ctx.AgentDB.Model(&models.AgentRegistration{}).
		Where("guid = ? AND state = ?", guid, models.AgentStatePending).
		Updates(map[string]interface{}{
			"state":         models.AgentStateRegistered,
			"registered_at": time.Now().UTC(),
		})

	if result.Error != nil {
		log.Errorf("POST /v1/admin/agents/%s/approve: failed: %v", guid, result.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "agent not found or not pending", http.StatusNotFound)
		return
	}

	log.Infof("POST /v1/admin/agents/%s/approve: agent approved", guid)
	w.WriteHeader(http.StatusOK)
}

// Deny handles POST /v1/admin/agents/{guid}/deny.
func Deny(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	guid := r.PathValue("guid")
	if guid == "" {
		http.Error(w, "guid required", http.StatusBadRequest)
		return
	}

	result := ctx.AgentDB.Where("guid = ? AND state = ?", guid, models.AgentStatePending).
		Delete(&models.AgentRegistration{})

	if result.Error != nil {
		log.Errorf("POST /v1/admin/agents/%s/deny: failed: %v", guid, result.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "agent not found or not pending", http.StatusNotFound)
		return
	}

	log.Infof("POST /v1/admin/agents/%s/deny: agent denied and removed", guid)
	w.WriteHeader(http.StatusOK)
}

// Revoke handles POST /v1/admin/agents/{guid}/revoke.
func Revoke(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	guid := r.PathValue("guid")
	if guid == "" {
		http.Error(w, "guid required", http.StatusBadRequest)
		return
	}

	result := ctx.AgentDB.Model(&models.AgentRegistration{}).
		Where("guid = ? AND state = ?", guid, models.AgentStateRegistered).
		Update("state", models.AgentStateRevoked)

	if result.Error != nil {
		log.Errorf("POST /v1/admin/agents/%s/revoke: failed: %v", guid, result.Error)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		http.Error(w, "agent not found or not registered", http.StatusNotFound)
		return
	}

	log.Infof("POST /v1/admin/agents/%s/revoke: agent revoked", guid)
	w.WriteHeader(http.StatusOK)
}
