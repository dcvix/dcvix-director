//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dcvix/dcvix-director/internal/token"
	log "github.com/sirupsen/logrus"
)

// RespondJSON sets the content type and encodes vit as JSON to the response writer.
func RespondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Errorf("Failed to encode JSON response: %v", err)
	}
}

// SetSessionCookie sets the dcvix_session cookie with the signed token.
func SetSessionCookie(w http.ResponseWriter, signedToken string) {
	expiration, err := token.GetExpiration(signedToken)
	if err != nil {
		log.Warnf("Failed to get session cookie token expiration: %v", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "dcvix_session",
		Value:    signedToken,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

// ExtractIP extracts the IP address from an address string (removes port).
func ExtractIP(remoteAddr string) string {
	lastColon := strings.LastIndex(remoteAddr, ":")
	if lastColon == -1 {
		return remoteAddr
	}
	ip := remoteAddr[:lastColon]
	ip = strings.TrimPrefix(ip, "[")
	ip = strings.TrimSuffix(ip, "]")
	return ip
}

// ClearSessionCookie clears the dcvix_session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "dcvix_session",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}
