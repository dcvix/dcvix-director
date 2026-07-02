//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package database

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/dcvix/dcvix-director/internal/models"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type userJSON struct {
	ID           string   `json:"id"`
	Admin        bool     `json:"admin,omitempty"`
	Workstations []string `json:"workstations,omitempty"`
	Pools        []string `json:"pools,omitempty"`
}

type poolJSON struct {
	ID           string   `json:"id"`
	Workstations []string `json:"workstations,omitempty"`
}

type PolicyDB struct {
	dirPath       string
	usersFilePath string
	poolsFilePath string

	db *gorm.DB
}

func NewPolicyDB(dirPath string) (*PolicyDB, error) {
	pdb := &PolicyDB{
		dirPath:       dirPath,
		usersFilePath: filepath.Join(dirPath, "users.json"),
		poolsFilePath: filepath.Join(dirPath, "pools.json"),
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Infof("Storage folder '%s' does not exist, creating folder", dirPath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage folder '%s': %w", dirPath, err)
		}
	}

	fileDefaults := []struct {
		path string
		def  []byte
	}{
		{pdb.usersFilePath, []byte("[]")},
		{pdb.poolsFilePath, []byte("[]")},
	}
	for _, f := range fileDefaults {
		if _, err := os.Stat(f.path); os.IsNotExist(err) {
			log.Infof("Data file '%s' does not exist, creating empty file", f.path)
			if err := os.WriteFile(f.path, f.def, 0644); err != nil {
				return nil, fmt.Errorf("failed to create data file '%s': %w", f.path, err)
			}
		}
	}

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open config database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&models.User{}, &models.UserWorkstation{}, &models.UserPool{},
		&models.Pool{}, &models.PoolWorkstation{}, &models.Workstation{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate config database: %w", err)
	}

	pdb.db = db

	if err := pdb.loadData(); err != nil {
		return nil, err
	}

	go pdb.watchForHUP()

	return pdb, nil
}

func (pdb *PolicyDB) loadData() error {
	usersBytes, err := os.ReadFile(pdb.usersFilePath)
	if err != nil {
		return fmt.Errorf("failed to read users file '%s': %w", pdb.usersFilePath, err)
	}
	var users []userJSON
	if err := json.Unmarshal(usersBytes, &users); err != nil {
		return fmt.Errorf("failed to parse users file '%s': %w", pdb.usersFilePath, err)
	}

	poolsBytes, err := os.ReadFile(pdb.poolsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read pools file '%s': %w", pdb.poolsFilePath, err)
	}
	var pools []poolJSON
	if err := json.Unmarshal(poolsBytes, &pools); err != nil {
		return fmt.Errorf("failed to parse pools file '%s': %w", pdb.poolsFilePath, err)
	}

	workstationSet := make(map[string]struct{})
	for _, u := range users {
		for _, ws := range u.Workstations {
			workstationSet[ws] = struct{}{}
		}
	}
	for _, p := range pools {
		for _, ws := range p.Workstations {
			workstationSet[ws] = struct{}{}
		}
	}

	tx := pdb.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	if err := tx.Exec("DELETE FROM pool_workstations").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM user_pools").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM user_workstations").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM pools").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM users").Error; err != nil {
		return err
	}
	if err := tx.Exec("DELETE FROM workstations").Error; err != nil {
		return err
	}

	for ws := range workstationSet {
		if err := tx.Create(&models.Workstation{Workstation: ws}).Error; err != nil {
			return fmt.Errorf("failed to insert workstation %s: %w", ws, err)
		}
	}

	for _, p := range pools {
		if err := tx.Create(&models.Pool{PoolID: p.ID}).Error; err != nil {
			return fmt.Errorf("failed to insert pool %s: %w", p.ID, err)
		}
		for _, ws := range p.Workstations {
			if err := tx.Create(&models.PoolWorkstation{
				PoolID:      p.ID,
				Workstation: ws,
			}).Error; err != nil {
				return fmt.Errorf("failed to insert pool_workstation %s/%s: %w", p.ID, ws, err)
			}
		}
	}

	for _, u := range users {
		if err := tx.Create(&models.User{
			UserID: u.ID,
			Admin:  u.Admin,
		}).Error; err != nil {
			return fmt.Errorf("failed to insert user %s: %w", u.ID, err)
		}
		for _, ws := range u.Workstations {
			if err := tx.Create(&models.UserWorkstation{
				UserID:      u.ID,
				Workstation: ws,
			}).Error; err != nil {
				return fmt.Errorf("failed to insert user_workstation %s/%s: %w", u.ID, ws, err)
			}
		}
		for _, p := range u.Pools {
			if err := tx.Create(&models.UserPool{
				UserID: u.ID,
				PoolID: p,
			}).Error; err != nil {
				return fmt.Errorf("failed to insert user_pool %s/%s: %w", u.ID, p, err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit config data: %w", err)
	}

	log.Infof("Successfully loaded users and pools from %s", pdb.dirPath)
	return nil
}

func (pdb *PolicyDB) watchForHUP() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	for {
		<-c
		log.Info("SIGHUP received. Reloading data...")
		if err := pdb.loadData(); err != nil {
			log.Errorf("Failed to reload data on SIGHUP: %v", err)
		}
	}
}

func (pdb *PolicyDB) GetUsers() []models.User {
	var users []models.User
	pdb.db.Preload("Workstations").Preload("Pools").Find(&users)
	return users
}

func (pdb *PolicyDB) GetPools() []models.Pool {
	var pools []models.Pool
	pdb.db.Preload("Workstations").Find(&pools)
	return pools
}

func (pdb *PolicyDB) GetAdmins() []string {
	var admins []string
	pdb.db.Model(&models.User{}).Where("admin = ?", true).Pluck("user_id", &admins)
	return admins
}

func (pdb *PolicyDB) GetServers(userID string) []string {
	var results []string
	pdb.db.Raw(`
		SELECT workstation FROM user_workstations WHERE user_id = ?
	`, userID).Scan(&results)
	if results == nil {
		return []string{}
	}
	return results
}

func (pdb *PolicyDB) GetPoolsForUser(userID string) []string {
	var results []string
	pdb.db.Raw(`
		SELECT pw.workstation FROM user_pools up
		JOIN pool_workstations pw ON up.pool_id = pw.pool_id
		WHERE up.user_id = ?
	`, userID).Scan(&results)
	if results == nil {
		return []string{}
	}
	return results
}

func (pdb *PolicyDB) IsAdmin(userID string) bool {
	var count int64
	pdb.db.Model(&models.User{}).Where("user_id = ? AND admin = ?", userID, true).Count(&count)
	return count > 0
}
