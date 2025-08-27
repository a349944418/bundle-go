package models

import (
	"time"

	"gorm.io/gorm"
)

type Shop struct {
	gorm.Model
	ShopDomain   string    `gorm:"uniqueIndex;not null"`
	AccessToken  string    `gorm:"not null"`
	RefreshToken string    // Optional
	Scope        string    `gorm:"not null"`
	ExpiresAt    time.Time // For online tokens
	InstalledAt  time.Time `gorm:"not null"`
}
