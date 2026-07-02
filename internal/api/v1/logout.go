//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package v1

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	log "github.com/sirupsen/logrus"
)

// Logout handles POST /v1/logout.
func Logout(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/logout: request from: %s", r.RemoteAddr)
	api.ClearSessionCookie(w)
}
