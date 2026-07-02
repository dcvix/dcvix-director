// SPDX-FileCopyrightText: 2025 Diego Cortassa
// SPDX-License-Identifier: MIT

package models

import "time"

// AgentUpdate is the payload sent by agents to the /v1/agent/update endpoint.
type AgentUpdate struct {
	Sessions []Session `json:"sessions"`
	Stats    struct {
		Hostname    string  `json:"hostname"`
		AgentIP     string  `json:"agent_ip"`
		FreeMemory  int64   `json:"free_memory"`
		TotalMemory int64   `json:"total_memory"`
		Cores       int     `json:"cores"`
		CPUUsage    float64 `json:"cpu_usage"`
		Load1       float64 `json:"load1"`
		Load5       float64 `json:"load5"`
		Load15      float64 `json:"load15"`
	} `json:"stats"`
	Tags []string `json:"tags"`
}

type AgentRegistrationState string

const (
	AgentStatePending    AgentRegistrationState = "pending"
	AgentStateRegistered AgentRegistrationState = "registered"
	AgentStateRevoked    AgentRegistrationState = "revoked"
)

type AgentRegistration struct {
	GUID         string                 `json:"guid" gorm:"primaryKey;column:guid"`
	Hostname     string                 `json:"hostname" gorm:"column:hostname"`
	State        AgentRegistrationState `json:"state" gorm:"column:state;not null;default:pending"`
	CreatedAt    time.Time              `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	RegisteredAt *time.Time             `json:"registered_at" gorm:"column:registered_at"`
	LastSeenAt   *time.Time             `json:"last_seen_at" gorm:"column:last_seen_at"`
	CertSerial   string                 `json:"cert_serial" gorm:"column:cert_serial"`
	Tags         []string               `json:"tags" gorm:"column:tags;serializer:json"`
}

func (AgentRegistration) TableName() string {
	return "agents"
}
