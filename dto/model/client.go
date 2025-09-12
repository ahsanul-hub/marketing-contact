package model

import (
	"encoding/json"
	"time"
)

type PaymentMethodClient struct {
	ID       uint            `gorm:"primaryKey" json:"id"`
	Name     string          `gorm:"size:255;not null" json:"name"`
	Route    json.RawMessage `gorm:"type:jsonb" json:"route"`
	Flexible bool            `json:"flexible"`
	Status   int             `json:"status"`
	Msisdn   int             `json:"msisdn"`
	ClientID string          `gorm:"size:50;not null" json:"client_id"`
	Fee      float64         `gorm:"type:decimal(10,2);default:0.00" json:"fee"`
}

type SettlementClient struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID          string    `gorm:"size:50;not null" json:"client_id"`
	Name              string    `gorm:"size:255;not null" json:"name"`
	IsBhpuso          string    `json:"is_bhpuso"`
	ServiceCharge     *string   `json:"servicecharge"`
	Tax23             *string   `json:"tax23"`
	Ppn               *float32  `json:"ppn"`
	Mdr               string    `json:"mdr"`
	MdrType           string    `json:"mdr_type"`
	AdditionalFee     *uint     `json:"additionalfee"`
	AdditionalPercent *float32  `json:"additional_percent"`
	AdditionalFeeType *int      `json:"additionalfee_type"`
	PaymentType       string    `json:"payment_type"`
	ShareRedision     *float32  `json:"share_redision"`
	SharePartner      *float32  `json:"share_partner"`
	IsDivide1Poin1    string    `json:"is_divide_1poin1"`
	UpdatedAt         time.Time `gorm:"not null" json:"updated_at"`
	CreatedAt         time.Time `gorm:"not null" json:"created_at"`
}

type Client struct {
	UID                string                `gorm:"size:50;primaryKey" json:"u_id"`
	ClientName         string                `gorm:"size:255;not null" json:"client_name"`
	ClientAppkey       string                `gorm:"size:255;not null;unique" json:"client_appkey"`
	ClientSecret       string                `gorm:"size:255;not null;unique" json:"client_secret"`
	ClientID           string                `gorm:"size:255;not null;unique" json:"client_appid"`
	AppName            string                `gorm:"size:255;not null" json:"app_name"`
	Address            string                `gorm:"size:255" json:"address"`
	Mobile             string                `gorm:"size:50;not null" json:"mobile"`
	ClientStatus       int                   `gorm:"not null" json:"client_status"`
	Phone              string                `gorm:"size:255" json:"phone"`
	Email              string                `gorm:"size:255" json:"email"`
	Testing            int                   `gorm:"size:10;not null" json:"testing"`
	Lang               string                `gorm:"size:10;not null" json:"lang"`
	CallbackURL        string                `gorm:"size:255;not null" json:"callback_url"`
	FailCallback       string                `gorm:"size:255" json:"fail_callback"`
	Isdcb              string                `gorm:"size:10;not null" json:"isdcb"`
	UpdatedAt          time.Time             `gorm:"not null" json:"updated_at"`
	CreatedAt          time.Time             `gorm:"not null" json:"created_at"`
	PaymentMethods     []PaymentMethodClient `gorm:"foreignKey:ClientID" json:"payment_methods"`
	Settlements        []SettlementClient    `gorm:"foreignKey:ClientID" json:"settlements"`
	ClientApps         []ClientApp           `gorm:"foreignKey:ClientID" json:"apps"`
	ChannelRouteWeight []ChannelRouteWeight  `gorm:"foreignKey:ClientID" json:"route_weights"`
}

type ClientApp struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AppName      string    `gorm:"size:255;not null" json:"app_name"`
	AppID        string    `gorm:"size:255;unique" json:"appid"`
	AppKey       string    `gorm:"size:255;unique" json:"appkey"`
	CallbackURL  string    `gorm:"size:255" json:"callback_url"`
	Testing      int       `gorm:"size:10;not null" json:"testing"`
	Status       int       `gorm:"not null" json:"status"`
	Mobile       string    `gorm:"size:20" json:"mobile"`
	FailCallback string    `gorm:"size:255" json:"fail_callback"`
	ClientID     string    `gorm:"not null" json:"client_id"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

type InputClientRequest struct {
	ClientName         *string               `json:"client_name"`
	AppName            *string               `json:"app_name"`
	Mobile             *string               `json:"mobile"`
	ClientStatus       *int                  `json:"client_status"`
	Testing            *int                  `json:"testing"`
	Lang               *string               `json:"lang"`
	Phone              *string               `json:"phone"`
	Email              *string               `json:"email"`
	Address            *string               `json:"address"`
	CallbackURL        *string               `json:"callback_url"`
	FailCallback       *string               `json:"fail_callback"`
	Isdcb              *string               `json:"isdcb"`
	PaymentMethods     []PaymentMethodClient `json:"payment_methods"`
	Settlements        []SettlementClient    `json:"settlements"`
	ClientApp          []ClientApp           `json:"client_app"`
	ChannelRouteWeight []ChannelRouteWeight  `json:"route_weights"`
}

type RouteWeight struct {
	Route  string  `json:"route"`
	Weight int     `json:"weight"`
	Fee    float64 `json:"fee"`
}

type SelectedPaymentMethod struct {
	PaymentMethodSlug string            `json:"payment_method_slug"`
	SelectedRoutes    []RouteWeight     `json:"selected_routes"`
	Status            int               `json:"status"`
	Msisdn            int               `json:"msisdn"`
	SettlementConfig  *SettlementConfig `json:"settlement_config,omitempty"`
}

type SettlementConfig struct {
	IsBhpuso          string   `json:"is_bhpuso"`
	ServiceCharge     *string  `json:"servicecharge"`
	Tax23             *string  `json:"tax23"`
	Ppn               *float32 `json:"ppn"`
	Mdr               string   `json:"mdr"`
	MdrType           string   `json:"mdr_type"`
	AdditionalFee     *uint    `json:"additionalfee"`
	AdditionalPercent *float32 `json:"additional_percent"`
	AdditionalFeeType *int     `json:"additionalfee_type"`
	PaymentType       string   `json:"payment_type"`
	ShareRedision     *float32 `json:"share_redision"`
	SharePartner      *float32 `json:"share_partner"`
	IsDivide1Poin1    string   `json:"is_divide_1poin1"`
}

type InputClientRequestV2 struct {
	ClientName             *string                 `json:"client_name"`
	AppName                *string                 `json:"app_name"`
	Mobile                 *string                 `json:"mobile,omitempty"`
	ClientStatus           *int                    `json:"client_status"`
	Testing                *int                    `json:"testing,omitempty"`
	Address                *string                 `json:"address"`
	Lang                   *string                 `json:"lang"`
	Phone                  *string                 `json:"phone"`
	Email                  *string                 `json:"email"`
	CallbackURL            *string                 `json:"callback_url"`
	FailCallback           *string                 `json:"fail_callback,omitempty"`
	Isdcb                  *string                 `json:"isdcb"`
	SelectedPaymentMethods []SelectedPaymentMethod `json:"selected_payment_methods"`
	ClientApp              []ClientApp             `json:"client_app"`
	// ChannelRouteWeight     []ChannelRouteWeight    `json:"route_weights"`
}

// ClientAppUpdate untuk update client app oleh client sendiri
type ClientAppUpdate struct {
	AppID        string  `json:"app_id"`
	CallbackURL  *string `json:"callback_url,omitempty"`
	FailCallback *string `json:"fail_callback,omitempty"`
	Mobile       *string `json:"mobile,omitempty"`
}
