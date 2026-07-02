//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package api

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/auth"
	"github.com/dcvix/dcvix-director/internal/ca"
	"github.com/dcvix/dcvix-director/internal/client"
	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/dcvix/dcvix-director/internal/database"
	"gorm.io/gorm"
)

// ContextKey is used for storing values in the request context.
type ContextKey string

const UserIDKey ContextKey = "userID"

// UserIDFromContext extracts the authenticated user ID from the request context.
func UserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// HandlerContext holds dependencies that handlers need.
type HandlerContext struct {
	Config        *config.Config
	MTLSClient    *client.MTLSClient
	DB            *gorm.DB
	Authenticator auth.Authenticator
	PolicyDB      *database.PolicyDB
	AgentDB       *gorm.DB
	Signer        *ca.Signer
}

// NewHandlerContext creates a new handler context with the given config.
func NewHandlerContext(cfg *config.Config, client *client.MTLSClient, db *gorm.DB, authenticator auth.Authenticator, policyDB *database.PolicyDB, agentDB *gorm.DB, signer *ca.Signer) *HandlerContext {
	return &HandlerContext{
		Config:        cfg,
		MTLSClient:    client,
		DB:            db,
		Authenticator: authenticator,
		PolicyDB:      policyDB,
		AgentDB:       agentDB,
		Signer:        signer,
	}
}

// HandlerFunc is a function that can handle HTTP requests with access to the handler context.
type HandlerFunc func(ctx *HandlerContext, w http.ResponseWriter, r *http.Request)

// ToHTTPHandlerFunc converts a HandlerFunc to an http.HandlerFunc.
func (ctx *HandlerContext) ToHTTPHandlerFunc(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(ctx, w, r)
	}
}
