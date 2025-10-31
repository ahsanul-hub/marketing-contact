package model

import "time"

type CreditCardLog struct {
	ID                              int64     `json:"id" db:"id"`
	PaymentSessionID                string    `json:"paymentSessionId" db:"payment_session_id"`
	PaymentSessionClientReferenceID string    `json:"paymentSessionClientReferenceId" db:"payment_session_client_reference_id"`
	StatementDescriptor             string    `json:"statementDescriptor" db:"statement_descriptor"`
	Status                          string    `json:"status" db:"status"`
	FailureCode                     string    `json:"failureCode" db:"failure_code"`
	FailureMessage                  string    `json:"failureMessage" db:"failure_message"`
	Recommendation                  string    `json:"recommendation" db:"recommendation"`
	CreatedAt                       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt                       time.Time `json:"updatedAt" db:"updated_at"`
	First6                          string    `json:"first6" db:"first6"`
	First8                          string    `json:"first8" db:"first8"`
	Last4                           string    `json:"last4" db:"last4"`
	ExpMonth                        string    `json:"expMonth" db:"exp_month"`
	ExpYear                         string    `json:"expYear" db:"exp_year"`
	CardType                        string    `json:"type" db:"card_type"`
	Brand                           string    `json:"brand" db:"brand"`
	IssuingBank                     string    `json:"issuingBank" db:"issuing_bank"`
	BinCountry                      string    `json:"country" db:"bin_country"`
	// AuthenticationResult
	ThreeDsVersion string `json:"threeDsVersion" db:"three_ds_version"`
	ThreeDsResult  string `json:"threeDsResult" db:"three_ds_result"`
	ThreeDsMethod  string `json:"threeDsMethod" db:"three_ds_method"`
	EciCode        string `json:"eciCode" db:"eci_code"`
	// AuthorizationResult
	AcquirerReferenceNumber  string `json:"acquirerReferenceNumber" db:"acquirer_reference_number"`
	RetrievalReferenceNumber string `json:"retrievalReferenceNumber" db:"retrieval_reference_number"`
	Stan                     string `json:"stan" db:"stan"`
	AvsResult                string `json:"avsResult" db:"avs_result"`
	CvvResult                string `json:"cvvResult" db:"cvv_result"`
	AuthorizedAmountValue    int    `json:"authorizedAmountValue" db:"authorized_amount_value"`
	AuthorizedAmountCurrency string `json:"authorizedAmountCurrency" db:"authorized_amount_currency"`
	IssuerAuthorizationCode  string `json:"issuerAuthorizationCode" db:"issuer_authorization_code"`
}
