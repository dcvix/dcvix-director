//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/dcvix/dcvix-director/internal/api"
	v1 "github.com/dcvix/dcvix-director/internal/api/v1"
	"github.com/dcvix/dcvix-director/internal/api/v1/admin"
	adminAgents "github.com/dcvix/dcvix-director/internal/api/v1/admin/agents"
	"github.com/dcvix/dcvix-director/internal/api/v1/agent"
	"github.com/dcvix/dcvix-director/internal/api/v1/servers"
	"github.com/dcvix/dcvix-director/internal/api/v1/sessions"
	"github.com/dcvix/dcvix-director/internal/api/v1/user"
)

// registerRoutes sets up all HTTP route handlers.
func (s *Server) registerRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	s.registerHealthRoutes(mux, handlerCtx)
	s.registerAuthRoutes(mux, handlerCtx)
	s.registerAdminRoutes(mux, handlerCtx)
	s.registerUserRoutes(mux, handlerCtx)
	s.registerAgentRoutes(mux, handlerCtx)
	s.registerGatewayRoutes(mux, handlerCtx)
}

func (s *Server) registerHealthRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("GET /v1/health", handlerCtx.ToHTTPHandlerFunc(v1.Health))
}

func (s *Server) registerAuthRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("GET /v1/auth-health", s.requireAuth(handlerCtx.ToHTTPHandlerFunc(v1.AuthHealth)))
	mux.HandleFunc("POST /v1/logout", handlerCtx.ToHTTPHandlerFunc(v1.Logout))
	mux.HandleFunc("POST /v1/admin/login", handlerCtx.ToHTTPHandlerFunc(admin.Login))
	mux.HandleFunc("POST /v1/user/login", handlerCtx.ToHTTPHandlerFunc(user.Login))
	mux.HandleFunc("POST /v1/user/authtokenverify", handlerCtx.ToHTTPHandlerFunc(user.AuthTokenVerify))
}

func (s *Server) registerAdminRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("GET /v1/sessions", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(sessions.List)))
	mux.HandleFunc("POST /v1/sessions", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(sessions.Create)))
	mux.HandleFunc("DELETE /v1/sessions/{id}", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(sessions.Close)))
	mux.HandleFunc("GET /v1/servers", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(servers.List)))
	mux.HandleFunc("GET /v1/admin/agents", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(adminAgents.ListPending)))
	mux.HandleFunc("POST /v1/admin/agents/{guid}/approve", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(adminAgents.Approve)))
	mux.HandleFunc("POST /v1/admin/agents/{guid}/deny", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(adminAgents.Deny)))
	mux.HandleFunc("POST /v1/admin/agents/{guid}/revoke", s.requireAdmin(handlerCtx.ToHTTPHandlerFunc(adminAgents.Revoke)))
}

func (s *Server) registerUserRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("GET /v1/user/servers", s.requireAuth(handlerCtx.ToHTTPHandlerFunc(user.Servers)))
	mux.HandleFunc("GET /v1/user/connectiontoken", s.requireAuth(handlerCtx.ToHTTPHandlerFunc(user.ConnectionToken)))
	mux.HandleFunc("POST /v1/user/sessions", s.requireAuth(handlerCtx.ToHTTPHandlerFunc(user.Create)))
	mux.HandleFunc("POST /v1/user/servers/{server}/config", s.requireAuth(handlerCtx.ToHTTPHandlerFunc(servers.SetConfig)))
}

func (s *Server) registerAgentRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("POST /v1/agent/update", s.requireClientCert(handlerCtx.ToHTTPHandlerFunc(agent.Update)))
	mux.HandleFunc("POST /v1/agent/register", handlerCtx.ToHTTPHandlerFunc(agent.Register))
	mux.HandleFunc("POST /v1/agent/renew", s.requireClientCert(handlerCtx.ToHTTPHandlerFunc(agent.Renew)))
}

func (s *Server) registerGatewayRoutes(mux *http.ServeMux, handlerCtx *api.HandlerContext) {
	mux.HandleFunc("POST /resolveSession", s.requireGatewayIP(handlerCtx.ToHTTPHandlerFunc(api.ResolveSession)))
}
