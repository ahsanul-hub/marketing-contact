package model

import "time"

type PaymentMethod struct {
	ID           uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug         string            `gorm:"type:VARCHAR(255);not null" json:"slug"`
	Description  string            `gorm:"type:TEXT;not null" json:"description"`
	Route        []string          `gorm:"type:TEXT" json:"route"`
	Type         string            `gorm:"type:VARCHAR(50);not null" json:"type"`
	Expired      string            `gorm:"type:TIMESTAMPTZ" json:"expired"`
	Report       string            `gorm:"type:TEXT;not null" json:"report"`
	JSONReturn   string            `gorm:"type:TEXT;not null" json:"json_return"`
	Parent       string            `gorm:"type:VARCHAR(50)" json:"parent"`
	IsAirtime    string            `gorm:"type:VARCHAR(10);not null" json:"is_airtime"`
	MinimumDenom float32           `gorm:"type:FLOAT;not null" json:"minimum_denom"`
	Disabled     string            `gorm:"type:VARCHAR(10);not null" json:"disabled"`
	Flexible     bool              `gorm:"not null" json:"flexible"`
	Status       string            `gorm:"type:VARCHAR(10);not null" json:"status"`
	Msisdn       string            `gorm:"type:VARCHAR(15)" json:"msisdn"`
	StatusDenom  map[string]string `gorm:"type:JSONB" json:"status_denom"`
	Prefix       []string          `gorm:"type:TEXT" json:"prefix"`
	Denom        []string          `gorm:"type:TEXT" json:"denom"`
	DailyLimit   int               `gorm:"not null" json:"daily_limit"`
	UpdatedAt    time.Time         `gorm:"not null" json:"updated_at"`
	CreatedAt    time.Time         `gorm:"not null" json:"created_at"`
}
