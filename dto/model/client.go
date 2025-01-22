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
	ShareRedision     *uint     `json:"share_redision"`
	SharePartner      *uint     `json:"share_partner"`
	IsDivide1Poin1    string    `json:"is_divide_1poin1"`
	UpdatedAt         time.Time `gorm:"not null" json:"updated_at"`
	CreatedAt         time.Time `gorm:"not null" json:"created_at"`
}

type Client struct {
	UID            string                `gorm:"size:50;primaryKey" json:"u_id"`
	ClientName     string                `gorm:"size:255;not null" json:"client_name"`
	ClientAppkey   string                `gorm:"size:255;not null;unique" json:"client_appkey"`
	ClientSecret   string                `gorm:"size:255;not null;unique" json:"client_secret"`
	ClientAppID    string                `gorm:"size:255;not null;unique" json:"client_appid"`
	AppName        string                `gorm:"size:255;not null" json:"app_name"`
	Mobile         string                `gorm:"size:50;not null" json:"mobile"`
	ClientStatus   int                   `gorm:"not null" json:"client_status"`
	Phone          string                `gorm:"size:255" json:"phone"`
	Email          string                `gorm:"size:255" json:"email"`
	Testing        int                   `gorm:"size:10;not null" json:"testing"`
	Lang           string                `gorm:"size:10;not null" json:"lang"`
	CallbackURL    string                `gorm:"size:255;not null" json:"callback_url"`
	FailCallback   string                `gorm:"size:10;not null" json:"fail_callback"`
	Isdcb          string                `gorm:"size:10;not null" json:"isdcb"`
	UpdatedAt      time.Time             `gorm:"not null" json:"updated_at"`
	CreatedAt      time.Time             `gorm:"not null" json:"created_at"`
	PaymentMethods []PaymentMethodClient `gorm:"foreignKey:ClientID" json:"payment_methods"`
	Settlements    []SettlementClient    `gorm:"foreignKey:ClientID" json:"settlements"`
}

type InputClientRequest struct {
	ClientName     *string               `json:"client_name"`
	AppName        *string               `json:"app_name"`
	Mobile         *string               `json:"mobile"`
	ClientStatus   *int                  `json:"client_status"`
	Testing        *int                  `json:"testing"`
	Lang           *string               `json:"lang"`
	CallbackURL    *string               `json:"callback_url"`
	FailCallback   *string               `json:"fail_callback"`
	Isdcb          *string               `json:"isdcb"`
	PaymentMethods []PaymentMethodClient `json:"payment_methods"`
	Settlements    []SettlementClient    `json:"settlements"`
}

// type PaymentMethodClient struct {
// 	Name   string                 `bson:"name" json:"name"`
// 	Route  map[string]interface{} `bson:"route" json:"route"`
// 	Status int                    `bson:"status" json:"status"`
// 	Msisdn int                    `bson:"msisdn" json:"msisdn"`
// }

// type Client struct {
// 	gorm.Model
// 	ID             string                `bson:"_id" json:"_id"`
// 	UID            string                `bson:"u_id" json:"u_id"`
// 	ClientName     string                `bson:"client_name" json:"client_name"`
// 	ClientAppkey   string                `bson:"client_appkey" json:"client_appkey"`
// 	ClientSecret   string                `bson:"client_secret" json:"client_secret"`
// 	ClientAppid    string                `bson:"client_appid" json:"client_appid"`
// 	AppName        string                `bson:"app_name" json:"app_name"`
// 	Mobile         string                `bson:"mobile" json:"mobile"`
// 	ClientStatus   int                   `bson:"client_status" json:"client_status"`
// 	Testing        string                `bson:"testing" json:"testing"`
// 	Lang           string                `bson:"lang" json:"lang"`
// 	CallbackURL    string                `bson:"callback_url" json:"callback_url"`
// 	PaymentMethods []PaymentMethodClient `gorm:"type:jsonb" bson:"payment_methods" json:"payment_methods"`
// 	FailCallback   string                `bson:"fail_callback" json:"fail_callback"`
// 	Isdcb          string                `bson:"isdcb" json:"isdcb"`
// 	UpdatedAt      time.Time             `bson:"updated_at" json:"updated_at"`
// 	CreatedAt      time.Time             `bson:"created_at" json:"created_at"`
// }
