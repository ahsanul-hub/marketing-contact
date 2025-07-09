package lib

import (
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

var HarsyaTokenCache = cache.New(10*time.Minute, 14*time.Minute)

type HarsyaTokenRequest struct {
	GrantType string `json:"grantType"`
}

type HarsyaTokenResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Data    TokenResponseData `json:"data"`
}

type TokenResponseData struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   string `json:"expiresIn"`
}

type VaChargingRequest struct {
	ClientReferenceID string `json:"clientReferenceId"`
	Amount            struct {
		Value    uint   `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"paymentMethod"`
	PaymentMethodOptions struct {
		VirtualAccount struct {
			Channel            string `json:"channel"`
			VirtualAccountName string `json:"virtualAccountName"`
		} `json:"virtualAccount"`
	} `json:"paymentMethodOptions"`
	Mode        string `json:"mode"`
	RedirectUrl struct {
		SuccessReturnUrl    string `json:"successReturnUrl"`
		FailureReturnUrl    string `json:"failureReturnUrl"`
		ExpirationReturnUrl string `json:"expirationReturnUrl"`
	} `json:"redirectUrl"`
	AutoConfirm bool   `json:"autoConfirm"`
	ExpiryAt    string `json:"expiryAt"`
}

// Struct responsenya bisa disesuaikan berdasarkan dokumentasi response dari API Harsya
// type VaChargingResponse struct {
// 	ID             string `json:"id"`
// 	Status         string `json:"status"`
// 	RedirectUrl    string `json:"redirectUrl"`
// 	InvoiceNo      string `json:"invoiceNo"`
// 	VirtualAccount struct {
// 		AccountNumber string `json:"accountNumber"`
// 		Bank          string `json:"bank"`
// 	} `json:"virtualAccount"`
// 	// Tambahkan field lain bila perlu
// }

type HarsyaChargingResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Data    VAChargingData `json:"data"`
}

type VAChargingData struct {
	ID                string `json:"id"`
	ClientReferenceID string `json:"clientReferenceId"`
	Amount            struct {
		Currency string `json:"currency"`
		Value    int    `json:"value"`
	} `json:"amount"`
	AutoConfirm   bool        `json:"autoConfirm"`
	Mode          string      `json:"mode"`
	RedirectURL   RedirectURL `json:"redirectUrl"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"paymentMethod"`
	StatementDesc      string         `json:"statementDescriptor"`
	Status             string         `json:"status"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	ExpiryAt           string         `json:"expiryAt"`
	PaymentURL         string         `json:"paymentUrl"`
	ChargeDetails      []ChargeDetail `json:"chargeDetails"`
	CancelledAt        *time.Time     `json:"cancelledAt"`
	CancellationReason *string        `json:"cancellationReason"`
	Metadata           Metadata       `json:"metadata"`
}

type RedirectURL struct {
	SuccessReturnURL    string `json:"successReturnUrl"`
	FailureReturnURL    string `json:"failureReturnUrl"`
	ExpirationReturnURL string `json:"expirationReturnUrl"`
}

type ChargeDetail struct {
	ID                        string `json:"id"`
	PaymentSessionID          string `json:"paymentSessionId"`
	PaymentSessionClientRefID string `json:"paymentSessionClientReferenceId"`
	Amount                    struct {
		Currency string `json:"currency"`
		Value    int    `json:"value"`
	} `json:"amount"`
	StatementDescriptor string         `json:"statementDescriptor"`
	Status              string         `json:"status"`
	AuthorizedAmount    *int           `json:"authorizedAmount,omitempty"`
	CapturedAmount      *int           `json:"capturedAmount,omitempty"`
	IsCaptured          bool           `json:"isCaptured"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	PaidAt              *time.Time     `json:"paidAt,omitempty"`
	VirtualAccount      VirtualAccount `json:"virtualAccount,omitempty"`
	Qr                  QrHarsya       `json:"qr,omitempty"`
}

type VirtualAccount struct {
	Channel              string    `json:"channel"`
	VirtualAccountNumber string    `json:"virtualAccountNumber"`
	VirtualAccountName   string    `json:"virtualAccountName"`
	ExpiryAt             time.Time `json:"expiryAt"`
}

type Metadata struct {
	InvoiceNo string `json:"invoiceNo"`
}

func RequestHarsyaAccessToken(clientID, clientSecret string) (*HarsyaTokenResponse, error) {
	// config, _ := config.GetGatewayConfig("xl_twt")
	// arrayOptions := config.Options["development"].(map[string]interface{})

	requestBody := HarsyaTokenRequest{
		GrantType: "client_credentials",
	}

	url := "https://api.harsya.com/v1/access-token"
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MERCHANT-ID", clientID)
	req.Header.Set("X-MERCHANT-SECRET", clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// body, err = ioutil.ReadAll(resp.Body)
	// if err != nil {

	// 	log.Println("Error reading response")
	// }

	// log.Println("res", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var tokenResp HarsyaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

func GetAccessTokenHarsya(clientID, clientSecret string) (string, error) {

	if cachedToken, found := HarsyaTokenCache.Get("accessToken"); found {
		// log.Println("Token diambil dari cache.")
		return cachedToken.(string), nil
	}

	tokenResp, err := RequestHarsyaAccessToken(clientID, clientSecret)
	if err != nil {
		return "", err
	}

	HarsyaTokenCache.Set("accessToken", tokenResp.Data.AccessToken, cache.DefaultExpiration)
	// log.Println("Token baru diminta dan disimpan ke cache.")

	return tokenResp.Data.AccessToken, nil
}

func VaHarsyaCharging(transactionId, customerName, bankName string, amount uint) (*HarsyaChargingResponse, error) {
	accessToken, err := GetAccessTokenHarsya("fd3bd903-ac6d-44e0-85cc-63435a4fb429", "P3J1PqOUlE8W1WpvUuKENzGTWQB1CXcbWGKWYkjt")
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	timeNow := time.Now().UTC()
	expiryAt := timeNow.Add(1 * time.Hour)

	requestBody := VaChargingRequest{
		ClientReferenceID: transactionId,
		Mode:              "REDIRECT",
		ExpiryAt:          expiryAt.Format(time.RFC3339),
		AutoConfirm:       true,
	}
	requestBody.Amount.Value = amount
	requestBody.Amount.Currency = "IDR"
	requestBody.PaymentMethod.Type = "VIRTUAL_ACCOUNT"
	requestBody.PaymentMethodOptions.VirtualAccount.Channel = bankName
	requestBody.PaymentMethodOptions.VirtualAccount.VirtualAccountName = customerName
	requestBody.RedirectUrl.SuccessReturnUrl = "https://merchant.com/success"
	requestBody.RedirectUrl.FailureReturnUrl = "https://merchant.com/failure"
	requestBody.RedirectUrl.ExpirationReturnUrl = "https://merchant.com/expiration"

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api-stg.harsya.com/v2/payments", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var chargingResp HarsyaChargingResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	now := time.Now()

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	err = repository.UpdateTransactionStatus(context.Background(), transactionId, 1001, &chargingResp.Data.ID, nil, "Processing payment", nil)
	if err != nil {
		log.Printf("Error updating transaction %s to PROCESSING: %s", transactionId, err)
	}

	return &chargingResp, nil
}
