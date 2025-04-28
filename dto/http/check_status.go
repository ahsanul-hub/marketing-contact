package http

import "time"

type TransactionStatus struct {
	UserID                string    `json:"user_id"`
	CreatedDate           time.Time `json:"created_date"`
	MerchantTransactionID string    `json:"merchant_transaction_id"`
	StatusCode            int       `json:"status_code,omitempty"`
	PaymentMethod         string    `json:"payment_method"`
	Amount                string    `json:"amount"`
	Status                string    `json:"status"`
	Currency              string    `json:"currency"`
	ItemName              string    `json:"item_name"`
	ItemID                string    `json:"item_id"`
	ReferenceID           string    `json:"reference_id"`
	AppID                 string    `json:"app_id,omitempty"`
	ClientAppKey          string    `json:"client_appkey,omitempty"`
}
