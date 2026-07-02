//  SPDX-FileCopyrightText: 2025 Diego Cortassa
//  SPDX-License-Identifier: MIT

package models

type Workstation struct {
	Workstation string `gorm:"primaryKey"`
}

type User struct {
	UserID       string            `gorm:"primaryKey"`
	Admin        bool              `gorm:"default:false;not null"`
	Workstations []UserWorkstation `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Pools        []UserPool        `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type UserWorkstation struct {
	UserID      string `gorm:"primaryKey"`
	Workstation string `gorm:"primaryKey"`
}

type UserPool struct {
	UserID string `gorm:"primaryKey"`
	PoolID string `gorm:"primaryKey"`
}

type Pool struct {
	PoolID       string            `gorm:"primaryKey"`
	Workstations []PoolWorkstation `gorm:"foreignKey:PoolID;constraint:OnDelete:CASCADE"`
}

type PoolWorkstation struct {
	PoolID      string `gorm:"primaryKey"`
	Workstation string `gorm:"primaryKey"`
}
