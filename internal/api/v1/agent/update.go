//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package agent

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

// Update handles POST /v1/agent/update.
func Update(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/agent/update: request from: %s", r.RemoteAddr)

	var update models.AgentUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx.DB.Where("server_id = ?", update.Stats.Hostname).Delete(&models.Session{})

	server := models.Server{
		Hostname:    update.Stats.Hostname,
		AgentIP:     api.ExtractIP(r.RemoteAddr),
		FreeMemory:  update.Stats.FreeMemory,
		TotalMemory: update.Stats.TotalMemory,
		Cores:       update.Stats.Cores,
		CPUUsage:    update.Stats.CPUUsage,
		Load1:       update.Stats.Load1,
		Load5:       update.Stats.Load5,
		Load15:      update.Stats.Load15,
		Tags:        update.Tags,
	}

	tx := ctx.DB.Begin()
	server.LastSeen = time.Now().UTC()
	if err := tx.Save(&server).Error; err != nil {
		tx.Rollback()
		log.Errorf("POST /v1/agent/update: failed to save server: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, session := range update.Sessions {
		session.UID = session.ID + "-" + server.Hostname
		session.ServerID = server.Hostname
		session.LastSeen = time.Now().UTC()
		if err := tx.Save(&session).Error; err != nil {
			tx.Rollback()
			log.Errorf("POST /v1/agent/update: failed to save session: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("POST /v1/agent/update: failed to commit transaction: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
