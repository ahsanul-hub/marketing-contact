package lib

import (
	"app/config"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type QrisHarsyaRequest struct {
	ClientReferenceID string `json:"clientReferenceId"`
	Amount            struct {
		Value    uint   `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"paymentMethod"`
	Mode        string `json:"mode"`
	RedirectUrl struct {
		SuccessReturnUrl    string `json:"successReturnUrl"`
		FailureReturnUrl    string `json:"failureReturnUrl"`
		ExpirationReturnUrl string `json:"expirationReturnUrl"`
	} `json:"redirectUrl"`
	AutoConfirm bool   `json:"autoConfirm"`
	ExpiryAt    string `json:"expiryAt"`
}

type QrHarsya struct {
	Acquirer                 string    `json:"acquirer"`
	QRContent                string    `json:"qrContent"`
	QRUrl                    string    `json:"qrUrl"`
	RetrievalReferenceNumber string    `json:"retrievalReferenceNumber"`
	IssuerName               string    `json:"issuerName"`
	ExpiryAt                 time.Time `json:"expiryAt"`
	MerchantName             string    `json:"merchantName,omitempty"`
}

func QrisHarsyaCharging(transactionId string, amount uint) (*HarsyaChargingResponse, error) {
	accessToken, err := GetAccessTokenHarsya("2e0ca65d-d5c2-4d55-8123-d049a888bce1", "IstUCDSYJDgbgCZsHu18xnGDesJjcJPRV3nZl4pN")
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	timeNow := time.Now().UTC()
	expiryAt := timeNow.Add(10 * time.Minute)

	successUrl := fmt.Sprintf("%s/return/dana", config.Config("APIURL", ""))

	requestBody := QrisHarsyaRequest{
		ClientReferenceID: transactionId,
		Mode:              "REDIRECT",
		ExpiryAt:          expiryAt.Format(time.RFC3339),
		AutoConfirm:       true,
	}

	requestBody.Amount.Value = amount
	requestBody.Amount.Currency = "IDR"
	requestBody.PaymentMethod.Type = "QR"

	requestBody.RedirectUrl.SuccessReturnUrl = successUrl
	requestBody.RedirectUrl.FailureReturnUrl = "https://merchant.com/failure"
	requestBody.RedirectUrl.ExpirationReturnUrl = "https://merchant.com/expiration"

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.harsya.com/v2/payments", bytes.NewBuffer(body))
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
