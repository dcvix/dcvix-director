// SPDX-FileCopyrightText: 2025 Diego Cortassa
// SPDX-License-Identifier: MIT

package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dcvix/dcvix-director/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitAgentDB opens or creates the persistent agents.db in dataDir.
// The file is created with 0600 permissions to protect registration data.
func InitAgentDB(dataDir string) (*gorm.DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "agents.db")

	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open agents.db: %w", err)
	}
	f.Close()

	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set agents.db permissions: %w", err)
	}

	gormConfig := gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=busy_timeout(500)&_pragma=journal_mode(WAL)"), &gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open agents database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.AgentRegistration{}); err != nil {
		return nil, fmt.Errorf("failed to migrate agents database: %w", err)
	}

	return db, nil
}
