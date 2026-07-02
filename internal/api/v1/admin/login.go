//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package admin

import (
	"encoding/json"
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/auth"
	"github.com/dcvix/dcvix-director/internal/token"
	log "github.com/sirupsen/logrus"
)

// Login handles POST /v1/admin/login.
func Login(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("POST /v1/admin/login: request from: %s", r.RemoteAddr)

	var user auth.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := ctx.Authenticator.Auth(user)
	if err != nil {
		log.Errorf("POST /v1/admin/login: authentication failed: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !response {
		log.Error("POST /v1/admin/login: authentication failed")
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	if !ctx.PolicyDB.IsAdmin(user.UserID) {
		log.Error("POST /v1/admin/login: user is not an admin")
		http.Error(w, "user is not an admin", http.StatusUnauthorized)
		return
	}

	signedToken := token.CreateToken(user.UserID, "admin")

	log.Debugf("POST /v1/admin/login: user %s logged in successfully", user.UserID)

	api.SetSessionCookie(w, signedToken)

	api.RespondJSON(w, map[string]string{})
}
