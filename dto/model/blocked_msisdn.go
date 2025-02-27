package model

import "time"

type BlockedMDN struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserMDN      string     `gorm:"type:VARCHAR(15);unique;not null" json:"user_mdn"`
	BlockedUntil *time.Time `gorm:"type:TIMESTAMP" json:"blocked_until"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
