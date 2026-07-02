//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package v1

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	log "github.com/sirupsen/logrus"
)

// AuthHealth handles GET /v1/auth-health (requires authentication).
func AuthHealth(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("GET /v1/auth-health: request from: %s", r.RemoteAddr)
	api.RespondJSON(w, map[string]string{"status": "healthy"})
}
