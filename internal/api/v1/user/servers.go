//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package user

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	log "github.com/sirupsen/logrus"
)

// Servers handles GET /v1/user/servers.
func Servers(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("GET /v1/user/servers: request from: %s", r.RemoteAddr)

	userID, ok := api.UserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Debugf("GET /v1/user/servers: user-ID from token: %s", userID)

	servers := ctx.PolicyDB.GetServers(userID)
	pools := ctx.PolicyDB.GetPoolsForUser(userID)
	// append pools to servers
	servers = append(servers, pools...)

	log.Debugf("GET /v1/user/servers: %v", servers)

	type ServersResponse struct {
		Servers []string `json:"servers"`
	}

	response := ServersResponse{
		Servers: []string{},
	}

	response.Servers = servers
	api.RespondJSON(w, response)
}
