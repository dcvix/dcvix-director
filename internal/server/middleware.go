//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package server

import (
	"context"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/token"
	log "github.com/sirupsen/logrus"
)

// requireAuth verifies the PASETO token allowing any user.
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("dcvix_session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := cookie.Value

		err = token.Verify(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := token.GetUserID(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if _, err := token.GetUserCategory(tokenString); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), api.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// requireAdmin verifies the PASETO token allowing only admin users.
func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("dcvix_session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := cookie.Value

		err = token.Verify(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := token.GetUserID(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if role, err := token.GetUserCategory(tokenString); err != nil || role != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), api.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// requireClientCert validates the mTLS client certificate against the CA pool.
func (s *Server) requireClientCert(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			clientCert := r.TLS.PeerCertificates[0]

			opts := x509.VerifyOptions{
				Roots:         s.caCertPool,
				CurrentTime:   time.Now(),
				Intermediates: x509.NewCertPool(),
			}

			for _, cert := range r.TLS.PeerCertificates[1:] {
				opts.Intermediates.AddCert(cert)
			}

			_, err := clientCert.Verify(opts)
			if err != nil {
				log.Errorf("Client certificate verification failed: %v from %v", err, r.RemoteAddr)
				http.Error(w, "Invalid client certificate", http.StatusUnauthorized)
				return
			}
		} else {
			log.Errorf("No client certificate provided from %v", r.RemoteAddr)
			http.Error(w, "No client certificate provided", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// requireGatewayIP restricts access to IPs in the gateway whitelist.
func (s *Server) requireGatewayIP(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gatewaysList := s.config.Gateway.GatewaysList
		if len(gatewaysList) == 0 {
			log.Warnf("Access denied from IP %s (gateways_list is empty)", api.ExtractIP(r.RemoteAddr))
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		clientIP := api.ExtractIP(r.RemoteAddr)
		for _, allowedIP := range gatewaysList {
			if clientIP == allowedIP {
				next.ServeHTTP(w, r)
				return
			}
		}

		log.Warnf("Access denied from IP %s (not in gateways_list)", clientIP)
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}
