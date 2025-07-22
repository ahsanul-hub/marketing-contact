package lib

import (
	"app/config"
	"app/repository"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Struct untuk address customer
type Address struct {
	Country string `json:"country"`
}

// Struct untuk customer
type Customer struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Address   Address `json:"address"`
}

// Struct untuk request body
type DigiphChargingRequest struct {
	ReferenceID     string   `json:"referenceId"`
	Description     string   `json:"description"`
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	CallbackURL     string   `json:"callbackUrl"`
	RedirectURL     string   `json:"redirectUrl"`
	InvoiceDuration int      `json:"invoiceDuration"`
	PaymentChannel  string   `json:"paymentChannel"`
	Customer        Customer `json:"customer"`
	Signature       string   `json:"signature"`
}

type DigiphChargingResponse struct {
	ID          string  `json:"id"`
	ReferenceID string  `json:"referenceId"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	ExpiresAt   string  `json:"expiresAt"`
	PaymentLink string  `json:"paymentLink"`
}

func generateSignature(req DigiphChargingRequest, secretKey string) string {
	concat := req.ReferenceID +
		fmt.Sprintf("%.0f", req.Amount) +
		req.Currency +
		fmt.Sprintf("%d", req.InvoiceDuration) +
		secretKey

	hash := sha256.Sum256([]byte(concat))
	return hex.EncodeToString(hash[:])
}

func RequestQrphCharging(transactionID, name, email string, amount uint) (*DigiphChargingResponse, error) {
	apiKey := "LTSQUlQsTOwbchFKoeliVnTBxr94TeUKPQIyYFKJ7ee04SrVXDuzfFJERzeUBaQI"
	username := "redision"
	password := "Pn12ifobQJPD7jam"
	secretKey := "0D9GazQOMRiP0D2d"

	reqBody := DigiphChargingRequest{
		ReferenceID:     transactionID,
		Description:     fmt.Sprintf("#order %s", transactionID),
		Amount:          float64(amount),
		Currency:        "PHP",
		CallbackURL:     fmt.Sprintf("%s/callback/digiph", config.Config("APIURL", "")),
		RedirectURL:     fmt.Sprintf("%s/return/dana", config.Config("APIURL", "")),
		InvoiceDuration: 3000,
		PaymentChannel:  "QRPH",
		Customer: Customer{
			FirstName: func() string {
				if name == "" {
					return "User"
				}
				return name
			}(),
			LastName: func() string {
				if name == "" {
					return "User"
				}
				return name
			}(),
			Email: email,
			Address: Address{
				Country: "PH",
			},
		},
	}

	reqBody.Signature = generateSignature(reqBody, secretKey)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("POST", "https://sandbox.payborit.com/api/checkout", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("API-Key", apiKey)
	request.Header.Set("Username", username)
	request.Header.Set("Password", password)

	// Execute
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("error response [%d]", resp.StatusCode)
	}

	var digiphResp DigiphChargingResponse
	if err := json.NewDecoder(resp.Body).Decode(&digiphResp); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	now := time.Now()

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	err = repository.UpdateTransactionStatus(context.Background(), transactionID, 1001, &digiphResp.ID, nil, "", nil)
	if err != nil {
		log.Printf("Error updating transaction %s status: %s", transactionID, err)
	}

	return &digiphResp, nil
}
