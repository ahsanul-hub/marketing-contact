package http

type CreatePaymentRequest struct {
	UserMDN               string  `json:"user_mdn" validate:"required"`
	UserID                string  `json:"user_id" validate:"required"`
	MerchantTransactionID string  `json:"merchant_transaction_id" validate:"required"`
	PaymentMethod         string  `json:"payment_method" validate:"required"`
	Amount                float32 `json:"amount" validate:"required,min=1"`
	ItemName              string  `json:"item_name" validate:"required,max=60"`
	Custom                string  `json:"custom,omitempty"`
}
