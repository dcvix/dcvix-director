//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package housekeeper

import (
	"context"
	"time"

	"github.com/dcvix/dcvix-director/internal/config"
	"github.com/dcvix/dcvix-director/internal/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Housekeeper struct {
	config    *config.Config
	runtimeDB *gorm.DB
	agentsDB  *gorm.DB
}

func NewHousekeeper(cfg *config.Config, runtimeDB *gorm.DB, agentsDB *gorm.DB) *Housekeeper {

	return &Housekeeper{
		config:    cfg,
		runtimeDB: runtimeDB,
		agentsDB:  agentsDB,
	}
}

func (h *Housekeeper) StartHousekeeper(ctx context.Context) {
	go h.runtimeVacuumLoop(ctx)
	go h.agentVacuumLoop(ctx)
}

func (h *Housekeeper) runtimeVacuumLoop(ctx context.Context) {
	freq, err := time.ParseDuration(h.config.Housekeeper.HousekeeperFrequency)
	if err != nil {
		log.Warnf("Invalid housekeeper_frequency '%s': %v, using default 40s", h.config.Housekeeper.HousekeeperFrequency, err)
		freq = 40 * time.Second
	}

	ticker := time.NewTicker(freq)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("Runtime database vacuum loop stopping")
			return
		case <-ticker.C:
			log.Debug("Runtime database vacuum loop running")

			if err := h.runtimeVacuum(); err != nil {
				log.Errorf("Failed vacuuming runtime database: %v", err)
			}
		}
	}
}

func (h *Housekeeper) agentVacuumLoop(ctx context.Context) {
	if h.agentsDB == nil {
		return
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("Agent database vacuum loop stopping")
			return
		case <-ticker.C:
			log.Debug("Agent database vacuum loop running")

			if err := h.agentVacuum(); err != nil {
				log.Errorf("Failed vacuuming agents database: %v", err)
			}
		}
	}
}

func (h *Housekeeper) runtimeVacuum() error {
	maxAge, err := time.ParseDuration(h.config.Housekeeper.MaxAge)
	if err != nil {
		log.Warnf("Invalid housekeeper max_age '%s': %v, using default 30s", h.config.Housekeeper.MaxAge, err)
		maxAge = 30 * time.Second
	}

	log.Debug("Vacuuming runtime database")

	if err := h.runtimeDB.Where("last_seen < ?", time.Now().UTC().Add(-maxAge)).Delete(&models.Server{}).Error; err != nil {
		return err
	}
	if err := h.runtimeDB.Where("last_seen < ?", time.Now().UTC().Add(-maxAge)).Delete(&models.Session{}).Error; err != nil {
		return err
	}

	return nil
}

func (h *Housekeeper) agentVacuum() error {
	log.Debug("Vacuuming agents database")
	// Remove registered agents not seen in 30 days
	if err := h.agentsDB.Where("state = ? AND last_seen_at < ?",
		"registered", time.Now().Add(-30*24*time.Hour)).Delete(&models.AgentRegistration{}).Error; err != nil {
		return err
	}

	// Remove pending agents older than 7 days
	if err := h.agentsDB.Where("state = ? AND created_at < ?",
		"pending", time.Now().Add(-7*24*time.Hour)).Delete(&models.AgentRegistration{}).Error; err != nil {
		return err
	}

	return nil
}
