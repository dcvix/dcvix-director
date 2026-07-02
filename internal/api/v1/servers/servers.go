//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package servers

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
)

// List handles GET /v1/servers.
func List(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("GET /v1/servers: request from: %s", r.RemoteAddr)

	var servers []models.Server
	if err := ctx.DB.Preload("Sessions").Order("hostname desc").Find(&servers).Error; err != nil {
		log.Errorf("GET /v1/servers: query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.RespondJSON(w, servers)
}
