//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package database

import (
	"fmt"

	"github.com/dcvix/dcvix-director/internal/models"
	// "gorm.io/driver/sqlite"   // Use the cgo GORM SQLite driver (cannot compile with CGO_ENABLED=0)
	"github.com/glebarez/sqlite" // Use the pure Go SQLite driver
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitRuntimeDB() (*gorm.DB, error) {
	var err error

	gormConfig := gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	// gormConfig := gorm.Config{Logger: logger.Default.LogMode(logger.Info)}

	db, err := gorm.Open(sqlite.Open("file::memory:?_pragma=busy_timeout(500)"), &gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// avoid database is locked error
	sqlDB.SetMaxOpenConns(1)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Server{}, &models.Session{}, &models.ConnectionToken{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate config database: %w", err)
	}

	return db, nil
}
