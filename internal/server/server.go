//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	stdliblog "log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dcvix/dcvix-director/frontend"
	"github.com/dcvix/dcvix-director/internal/api"
	"github.com/dcvix/dcvix-director/internal/auth"
	"github.com/dcvix/dcvix-director/internal/ca"
	"github.com/dcvix/dcvix-director/internal/client"
	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/dcvix/dcvix-director/internal/database"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Server holds the HTTP server configuration and dependencies.
type Server struct {
	config        *config.Config
	db            *gorm.DB
	agentDB       *gorm.DB
	authenticator auth.Authenticator
	policyDB      *database.PolicyDB
	signer        *ca.Signer
	httpServer    *http.Server
	caCertPool    *x509.CertPool
}

// NewServer creates a new Server instance.
func NewServer(cfg *config.Config, db *gorm.DB, agentDB *gorm.DB, authenticator auth.Authenticator, policyDB *database.PolicyDB, signer *ca.Signer) *Server {
	return &Server{
		config:        cfg,
		db:            db,
		agentDB:       agentDB,
		authenticator: authenticator,
		policyDB:      policyDB,
		signer:        signer,
		caCertPool:    signer.CACertPool(),
	}
}

// Start builds the mux, registers routes, and starts the HTTPS listener.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	mTLSClient, err := client.NewMTLSClient(
		filepath.Join(s.config.Director.DataDir, "server.crt"),
		filepath.Join(s.config.Director.DataDir, "server.key"),
		filepath.Join(s.config.Director.DataDir, "ca.pem"),
	)
	if err != nil {
		return fmt.Errorf("failed to create mTLS client: %w", err)
	}

	handlerCtx := api.NewHandlerContext(s.config, mTLSClient, s.db, s.authenticator, s.policyDB, s.agentDB, s.signer)

	s.registerRoutes(mux, handlerCtx)

	s.registerStaticFiles(mux)

	tlsConfig := &tls.Config{
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  s.caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Director.DirectorHost, s.config.Director.DirectorPort),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ErrorLog:     stdliblog.New(log.StandardLogger().WriterLevel(log.WarnLevel), "", 0),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	certFile := filepath.Join(s.config.Director.DataDir, "server.crt")
	keyFile := filepath.Join(s.config.Director.DataDir, "server.key")
	log.Infof("Listening on https://%s:%d", s.config.Director.DirectorHost, s.config.Director.DirectorPort)
	return s.httpServer.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// registerStaticFiles serves the frontend static files with SPA fallback.
func (s *Server) registerStaticFiles(mux *http.ServeMux) {
	var distDir = "./frontend/dist/"
	var hfs http.Handler
	var staticFS fs.FS
	if stat, err := os.Stat(distDir); err == nil && stat.IsDir() {
		log.Info("Serving static files from ./frontend/dist directory")
		staticFS = os.DirFS(distDir)
		hfs = http.FileServer(http.Dir(distDir))
	} else {
		log.Info("Serving static files from embedded files")
		subFS, err := fs.Sub(frontend.EmbeddedFiles, "dist")
		if err != nil {
			log.Fatalf("could not create sub-filesystem: %v", err)
		}
		staticFS = subFS
		hfs = http.FileServer(http.FS(subFS))
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(staticFS, path); os.IsNotExist(err) {
			indexFile, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(indexFile)
			return
		}
		hfs.ServeHTTP(w, r)
	})
}
