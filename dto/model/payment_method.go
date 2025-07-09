package model

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type PaymentMethod struct {
	ID           uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug         string          `gorm:"type:VARCHAR(255);not null" json:"slug"`
	Description  string          `gorm:"type:TEXT" json:"description"`
	Route        pq.StringArray  `gorm:"type:TEXT" json:"route"`
	Type         string          `gorm:"type:VARCHAR(50);not null" json:"type"`
	Expired      string          `gorm:"type:VARCHAR(50)" json:"expired"`
	Report       string          `gorm:"type:TEXT" json:"report"`
	JSONReturn   string          `gorm:"type:TEXT" json:"json_return"`
	Parent       string          `gorm:"type:VARCHAR(50)" json:"parent"`
	IsAirtime    string          `gorm:"type:VARCHAR(10)" json:"is_airtime"`
	MinimumDenom float32         `gorm:"type:INTEGER;not null" json:"minimum_denom"`
	Disabled     string          `gorm:"type:VARCHAR(10)" json:"disabled"`
	Flexible     bool            `gorm:"not null" json:"flexible"`
	Status       string          `gorm:"type:VARCHAR(10);not null" json:"status"`
	Msisdn       string          `gorm:"type:VARCHAR(15)" json:"msisdn"`
	StatusDenom  json.RawMessage `gorm:"type:JSONB" json:"status_denom"`
	Prefix       pq.StringArray  `gorm:"type:TEXT" json:"prefix"`
	Denom        pq.StringArray  `gorm:"type:TEXT" json:"denom"`
	DailyLimit   int             `json:"daily_limit"`
	UpdatedAt    time.Time       `gorm:"not null" json:"updated_at"`
	CreatedAt    time.Time       `gorm:"not null" json:"created_at"`
}

type ChannelRouteWeight struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClientID      string    `gorm:"not null" json:"client_id"`
	PaymentMethod string    `gorm:"not null" json:"payment_method"`
	Route         string    `gorm:"not null" json:"route"`
	Weight        int       `gorm:"not null" json:"weight"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
