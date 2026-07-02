//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package user

import (
	"net/http"
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/models"
	"github.com/dcvix/dcvix-director/internal/token"
	log "github.com/sirupsen/logrus"
)

// ConnectionToken handles GET /v1/user/connectiontoken.
func ConnectionToken(ctx *api.HandlerContext, w http.ResponseWriter, r *http.Request) {
	log.Debugf("GET /v1/user/connectiontoken: request from: %s", r.RemoteAddr)

	userID, ok := api.UserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	SessionID := r.FormValue("sessionId")
	ServerID := r.FormValue("serverId")

	log.Debugf("GET /v1/user/connectiontoken: request for userID: %s sessionID: %s serverID: %s", userID, SessionID, ServerID)

	connectionToken := models.ConnectionToken{}

	err := ctx.DB.First(&connectionToken, "id = ?", userID+"-"+ServerID+"-"+SessionID).Error
	log.Debugf("GET /v1/user/connectiontoken: %+v", connectionToken)

	if err != nil {
		log.Debugf("GET /v1/user/connectiontoken: not found for user: %s, server: %s, session: %s", userID, ServerID, SessionID)
	}

	var signedToken string
	if connectionToken.Token != "" {
		expires, err := token.GetExpiration(connectionToken.Token)
		if err == nil && !time.Now().After(expires) {
			signedToken = connectionToken.Token
		}
	}

	if signedToken == "" {
		log.Debug("GET /v1/user/connectiontoken: creating new connection token")
		signedToken = token.CreateToken(userID, "user")
		connectionToken.ID = userID + "-" + ServerID + "-" + SessionID
		connectionToken.UserID = userID
		connectionToken.ServerID = ServerID
		connectionToken.SessionID = SessionID
		connectionToken.Token = signedToken
		expires, err := token.GetExpiration(signedToken)
		if err != nil {
			log.Errorf("GET /v1/user/connectiontoken: failed to get expiration: %v", err)
			http.Error(w, "Failed to create connection token", http.StatusInternalServerError)
			return
		}
		connectionToken.Expires = expires
		ctx.DB.Save(&connectionToken)
	}

	api.RespondJSON(w, map[string]string{
		"token": signedToken,
	})
}
