//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/dcvix/dcvix-director/internal/auth"
	"github.com/dcvix/dcvix-director/internal/ca"
	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/dcvix/dcvix-director/internal/database"
	"github.com/dcvix/dcvix-director/internal/housekeeper"
	"github.com/dcvix/dcvix-director/internal/logger"
	"github.com/dcvix/dcvix-director/internal/models"
	"github.com/dcvix/dcvix-director/internal/server"
	"github.com/dcvix/dcvix-director/internal/token"
	"github.com/dcvix/dcvix-director/internal/version"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	configPath := flag.String("conf", "", "Path to the configuration file")
	approveAgent := flag.String("approve-agent", "", "Approve a pending agent by GUID")
	revokeAgent := flag.String("revoke-agent", "", "Revoke a registered agent by GUID")
	listPending := flag.Bool("list-pending-agents", false, "List all pending agents")
	flag.Parse()

	if *showVersion {
		fmt.Println("dcvix-director version:", version.String())
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("Unable to determine version information.")
			return
		}
		fmt.Println("buildInfo:", buildInfo.Main.Version)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := logger.SetupLogger(cfg.Log); err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}

	log.Infof("Version: %s", version.String())
	log.Info("Starting")
	log.Infof("Configuration file: %s", *configPath)

	if err := token.InitSymmetricKey(cfg.Director.TokenKey); err != nil {
		log.Fatalf("Failed to initialize token key: %v", err)
	}

	var authenticator auth.Authenticator
	switch cfg.Director.AuthType {
	case "external":
		authenticator = auth.NewExternalAuthenticator(cfg.AuthExternal)
	case "ldap":
		authenticator = auth.NewLDAPAuthenticator(cfg.AuthLDAP)
	case "radius":
		authenticator = auth.NewRadiusAuthenticator(cfg.AuthRadius)
	default:
		authenticator = auth.NewPamAuthenticator(cfg.AuthPAM)
	}
	log.Infof("Authenticator initialized: %s", cfg.Director.AuthType)

	policyDB, err := database.NewPolicyDB(cfg.Director.PolicyDBFolder)
	if err != nil {
		log.Fatalf("Failed to initialize policy database: %v", err)
	}

	if *listPending || *approveAgent != "" || *revokeAgent != "" {
		agentDB, err := database.InitAgentDB(cfg.Director.DataDir)
		if err != nil {
			log.Fatalf("Failed to initialize agents database: %v", err)
		}
		runAgentCommand(agentDB, *listPending, *approveAgent, *revokeAgent)
		os.Exit(0)
	}

	runtimeDB, err := database.InitRuntimeDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Info("Runtime database initialized")

	agentDB, err := database.InitAgentDB(cfg.Director.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize agents database: %v", err)
	}
	log.Info("Agents database initialized")

	signer, err := ca.NewSigner(cfg.Director.DataDir)
	if err != nil {
		log.Fatalf("Failed to initialize CA: %v", err)
	}
	log.Info("CA initialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := housekeeper.NewHousekeeper(cfg, runtimeDB, agentDB)
	go h.StartHousekeeper(ctx)
	log.Info("Housekeeper started")

	s := server.NewServer(cfg, runtimeDB, agentDB, authenticator, policyDB, signer)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Server error: %v", err)
			sigChan <- syscall.SIGTERM
		}
	}()

	<-sigChan
	log.Info("Shutting down")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Errorf("Server shutdown error: %v", err)
	}

	sqlRuntimeDB, err := runtimeDB.DB()
	if err == nil {
		if err := sqlRuntimeDB.Close(); err != nil {
			log.Errorf("Runtime database close error: %v", err)
		}
	}

	sqlAgentDB, err := agentDB.DB()
	if err == nil {
		if err := sqlAgentDB.Close(); err != nil {
			log.Errorf("Agents database close error: %v", err)
		}
	}

	log.Info("Shutdown complete")
}

func runAgentCommand(agentDB *gorm.DB, listPending bool, approveGUID, revokeGUID string) {
	if listPending {
		var agents []models.AgentRegistration
		if err := agentDB.Where("state = ?", models.AgentStatePending).Find(&agents).Error; err != nil {
			log.Fatalf("Failed to list pending agents: %v", err)
		}
		if len(agents) == 0 {
			fmt.Println("No pending agents")
			return
		}
		for _, a := range agents {
			fmt.Printf("%s  %s  %s\n", a.GUID, a.Hostname, a.CreatedAt.Format(time.RFC3339))
		}
		return
	}
	if approveGUID != "" {
		result := agentDB.Model(&models.AgentRegistration{}).
			Where("guid = ? AND state = ?", approveGUID, models.AgentStatePending).
			Updates(map[string]interface{}{
				"state":         models.AgentStateRegistered,
				"registered_at": time.Now().UTC(),
			})
		if result.Error != nil {
			log.Fatalf("Failed to approve agent %s: %v", approveGUID, result.Error)
		}
		if result.RowsAffected == 0 {
			log.Fatalf("Agent %s not found or not pending", approveGUID)
		}
		fmt.Printf("Agent %s approved\n", approveGUID)
		return
	}
	if revokeGUID != "" {
		result := agentDB.Model(&models.AgentRegistration{}).
			Where("guid = ? AND state = ?", revokeGUID, models.AgentStateRegistered).
			Update("state", models.AgentStateRevoked)
		if result.Error != nil {
			log.Fatalf("Failed to revoke agent %s: %v", revokeGUID, result.Error)
		}
		if result.RowsAffected == 0 {
			log.Fatalf("Agent %s not found or not registered", revokeGUID)
		}
		fmt.Printf("Agent %s revoked\n", revokeGUID)
	}
}
