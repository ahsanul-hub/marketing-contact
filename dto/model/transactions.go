package model

import (
	"time"
)

type Transactions struct {
	ID                       string     `gorm:"size:50;primaryKey" json:"u_id"`
	BersamaBookingId         string     `gorm:"type:VARCHAR(255);not null" json:"bersama_booking_id"`
	SmsCode                  string     `gorm:"type:VARCHAR(255)"`
	MerchantName             string     `gorm:"type:VARCHAR(255);not null" json:"merchant_name"`
	AppName                  string     `gorm:"type:VARCHAR(255)" json:"app_name"`
	Keyword                  string     `gorm:"type:VARCHAR(255)"`
	Otp                      int        `json:"otp"`
	TcashId                  string     `gorm:"type:VARCHAR(255)" json:"tcach_id"`
	VaBcadynamicFaspayBillno int        `gorm:"type:INTEGER" json:"va_bca"`
	MtTid                    string     `gorm:"type:VARCHAR(255)" json:"merchant_transaction_id"`
	DisbursementId           string     `gorm:"type:VARCHAR(255)" json:"disbursement_id"`
	PaymentMethod            string     `gorm:"type:VARCHAR(255)" json:"payment_method"`
	StatusCode               int        `gorm:"type:INTEGER" json:"status_code"`
	ItemName                 string     `gorm:"type:VARCHAR(255);not null" json:"item_name"`
	ItemId                   string     `gorm:"type:VARCHAR(255)" json:"item_id"`
	Route                    string     `gorm:"type:TEXT" json:"route"`
	MdmTrxID                 string     `gorm:"type:VARCHAR(255)" json:"mdm_trx_id"`
	UserId                   string     `gorm:"type:VARCHAR(255)" json:"user_id"`
	TimestampRequestDate     *time.Time `json:"timestamp_request_date"`
	TimestampSubmitDate      *time.Time `json:"timestamp_submit_date"`
	TimestampCallbackDate    *time.Time `json:"timestamp_callback_date"`
	TimestampCallbackResult  string     `gorm:"type:VARCHAR(255)" json:"timestamp_callback_result"`
	Stan                     string     `gorm:"type:VARCHAR(255)" json:"json"`
	Amount                   uint       `gorm:"type:INTEGER" json:"amount"`
	ClientAppKey             string     `gorm:"type:VARCHAR(255)" json:"client_appkey"`
	AppID                    string     `gorm:"type:VARCHAR(255)" json:"appid"`
	Testing                  bool       `gorm:"type:BOOLEAN" json:"testing"`
	Token                    string     `gorm:"type:VARCHAR(255)" json:"token"`
	Currency                 string     `gorm:"type:VARCHAR(10)" json:"currency"`
	NetSettlement            float32    `gorm:"type:INTEGER" json:"net_settlement"`
	Price                    uint       `gorm:"type:INTEGER" json:"price"`
	BodySign                 string     `gorm:"type:TEXT" json:"bodysign"`
	UserMDN                  string     `gorm:"type:VARCHAR(15)" json:"user_mdn"`
	RedirectURL              string     `gorm:"type:TEXT" json:"redirect_url"`
	RedirectTarget           string     `gorm:"type:TEXT" json:"redirect_target"`
	CreatedAt                time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time  `gorm:"autoCreateTime" json:"updated_at"`
}

type TransactionMerchantResponse struct {
	UserMDN                 string     `json:"user_mdn"`
	UserID                  string     `json:"user_id"`
	PaymentMethod           string     `json:"payment_method"`
	MerchantTransactionID   string     `json:"merchant_transaction_id"`
	StatusCode              int        `json:"status_code"`
	TimestampRequestDate    *time.Time `json:"timestamp_request_date"`
	TimestampSubmitDate     *time.Time `json:"timestamp_submit_date"`
	TimestampCallbackDate   *time.Time `json:"timestamp_callback_date"`
	TimestampCallbackResult string     `json:"timestamp_callback_result"`
	ItemId                  string     `json:"item_id"`
	ItemName                string     `json:"item_name"`
	Route                   string     `json:"route"`
	Amount                  uint       `json:"amount"`
	Price                   uint       `json:"price"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type InputPaymentRequest struct {
	ClientAppKey  string `json:"client_appkey"`
	AppName       string `json:"app_name"`
	AppID         string `json:"app_id"`
	Status        string `json:"status"`
	MtTid         string `json:"merchant_transaction_id"`
	ItemId        string `json:"item_id"`
	Mobile        string `json:"mobile"`
	Testing       bool   `json:"testing"`
	Route         string `json:"route"`
	PaymentMethod string `json:"payment_method"`
	UserId        string `json:"user_id"`
	Currency      string `json:"currency"`
	Price         uint   `json:"price"`
	Amount        uint   `json:"amount"`
	ItemName      string `json:"item_name"`
	UserMDN       string `json:"user_mdn"`
	RedirectURL   string `json:"redirect_url"`
}
