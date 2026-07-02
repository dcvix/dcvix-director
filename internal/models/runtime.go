//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package models

import (
	"time"
)

// StatusClosing is set by the director when a session close has been requested
// but the agent hasn't confirmed it yet.
const StatusClosing = "closing"

// type Session struct {
// 	ID           string    `json:"id" gorm:"primaryKey"`
// 	Owner        string    `json:"owner"`
// 	User         string    `json:"user"`
// 	CreationTime time.Time `json:"creation-time"`
// 	State        string    `json:"state"`
// 	Type         string    `json:"type"`
// 	ServerID     string    `json:"server_id"`
// }

type Session struct {
	UID      string    `json:"uid" gorm:"primaryKey"` // uniq id Session ID + Server ID
	ServerID string    `json:"server_id"`
	LastSeen time.Time `json:"last-seen" gorm:"index"`

	// From agent update
	ID                    string    `json:"id"`
	Owner                 string    `json:"owner"`
	User                  string    `json:"user"`
	NumOfConnections      int       `json:"num-of-connections"`
	CreationTime          time.Time `json:"creation-time"`
	LastDisconnectionTime time.Time `json:"last-disconnection-time"`
	// Licenses              []struct {
	// 	Product        string    `json:"product"`
	// 	Status         string    `json:"status"`
	// 	CheckTimestamp time.Time `json:"check-timestamp"`
	// 	ExpirationDate time.Time `json:"expiration-date"`
	// } `json:"licenses"`
	LicensingMode string `json:"licensing-mode"`
	StorageRoot   string `json:"storage-root"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	X11Display    string `json:"x11-display"`
	X11Authority  string `json:"x11-authority"`
	// DisplayLayout []struct {
	// 	Width  int `json:"width"`
	// 	Height int `json:"height"`
	// 	X      int `json:"x"`
	// 	Y      int `json:"y"`
	// } `json:"display-layout"`
}

type Server struct {
	LastSeen time.Time `json:"last-seen" gorm:"index"`

	// From agent update
	Hostname    string    `json:"hostname" gorm:"primaryKey"`
	AgentIP     string    `json:"agent_ip"`
	FreeMemory  int64     `json:"free_memory"`
	TotalMemory int64     `json:"total_memory"`
	Cores       int       `json:"cores"`
	CPUUsage    float64   `json:"cpu_usage"`
	Load1       float64   `json:"load1"`
	Load5       float64   `json:"load5"`
	Load15      float64   `json:"load15"`
	Tags        []string  `json:"tags" gorm:"serializer:json"`
	Sessions    []Session `json:"sessions" gorm:"foreignKey:ServerID"`
}

type ConnectionToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	ServerID  string    `json:"server_id"`
	SessionID string    `json:"session_id"`
	Token     string    `json:"token"`
	Expires   time.Time `json:"expires" gorm:"index"`
}
