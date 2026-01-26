package lib

import (
	"app/config"
	"app/dto/model"
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DokuCheckoutRequest represents the request payload for DOKU checkout
type DokuCheckoutRequest struct {
	Order struct {
		Amount              uint   `json:"amount"`
		InvoiceNumber       string `json:"invoice_number"`
		Currency            string `json:"currency"`
		CallbackURL         string `json:"callback_url,omitempty"`
		CallbackURLCancel   string `json:"callback_url_cancel,omitempty"`
		CallbackURLResult   string `json:"callback_url_result,omitempty"`
		Language            string `json:"language"`
		AutoRedirect        bool   `json:"auto_redirect"`
		DisableRetryPayment bool   `json:"disable_retry_payment,omitempty"`
		LineItems           []struct {
			Name  string `json:"name"`
			Price uint   `json:"price"`
		} `json:"line_items,omitempty"`
	} `json:"order"`
	Payment struct {
		PaymentDueDate int `json:"payment_due_date"`
		// Type               string   `json:"type"`
		PaymentMethodTypes []string `json:"payment_method_types"`
	} `json:"payment"`
	Customer struct {
		ID       string `json:"id,omitempty"`
		Name     string `json:"name"`
		LastName string `json:"last_name,omitempty"`
		Phone    string `json:"phone,omitempty"`
		Email    string `json:"email,omitempty"`
		Address  string `json:"address,omitempty"`
		Postcode string `json:"postcode,omitempty"`
		State    string `json:"state,omitempty"`
		City     string `json:"city,omitempty"`
		Country  string `json:"country,omitempty"`
	} `json:"customer"`
}

// DokuCheckoutResponse represents the response from DOKU checkout API
type DokuCheckoutResponse struct {
	Message  []string `json:"message"`
	Response struct {
		Order struct {
			Amount        string `json:"amount"`
			InvoiceNumber string `json:"invoice_number"`
			Currency      string `json:"currency"`
			SessionID     string `json:"session_id"`
		} `json:"order"`
		Payment struct {
			PaymentMethodTypes []string `json:"payment_method_types"`
			PaymentDueDate     int      `json:"payment_due_date"`
			TokenID            string   `json:"token_id"`
			URL                string   `json:"url"`
			ExpiredDate        string   `json:"expired_date"`
			ExpiredDatetime    string   `json:"expired_datetime"`
			Type               string   `json:"type"`
		} `json:"payment"`
		Customer struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			LastName string `json:"last_name"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
		} `json:"customer"`
	} `json:"response"`
}

// generateDigest calculates SHA256 + Base64 from body JSON (for POST requests)
func generateDigest(body string) string {
	hash := sha256.Sum256([]byte(body))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// generateRequestSignature creates signature from request components
func generateRequestSignature(
	clientID, requestID, requestTimestamp, requestTarget, digest, secret string,
) string {
	// Build component string in order:
	// Client-Id, Request-Id, Request-Timestamp, Request-Target, Digest (if present)
	var sb strings.Builder
	sb.WriteString("Client-Id:" + clientID + "\n")
	sb.WriteString("Request-Id:" + requestID + "\n")
	sb.WriteString("Request-Timestamp:" + requestTimestamp + "\n")
	sb.WriteString("Request-Target:" + requestTarget)
	if len(digest) > 0 {
		sb.WriteString("\n")
		sb.WriteString("Digest:" + digest)
	}

	// Calculate HMAC-SHA256 with secret key
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sb.String()))

	// Encode to base64 and add prefix "HMACSHA256="
	return "HMACSHA256=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// RequestDokuCheckout creates a DOKU checkout payment
func RequestDokuCheckout(transactionID string, amount uint, customerInfo model.InputPaymentRequest) (*DokuCheckoutResponse, error) {
	clientID := config.Config("DOKU_CLIENT_ID", "")
	secretKey := config.Config("DOKU_SECRET_KEY", "")

	baseURL := config.Config("DOKU_BASE_URL", "https://api-sandbox.doku.com")

	if clientID == "" || secretKey == "" {
		return nil, fmt.Errorf("DOKU credentials not configured")
	}

	// Build request payload
	chargeRequest := DokuCheckoutRequest{}
	chargeRequest.Order.Amount = amount
	chargeRequest.Order.InvoiceNumber = transactionID
	chargeRequest.Order.Currency = "IDR"
	chargeRequest.Order.Language = "EN"
	// chargeRequest.Order.AutoRedirect = true
	chargeRequest.Order.DisableRetryPayment = true

	// Set callback URLs
	apiURL := config.Config("APIURL", "")
	if customerInfo.NotificationUrl != "" {
		chargeRequest.Order.CallbackURL = customerInfo.NotificationUrl
	} else {
		chargeRequest.Order.CallbackURL = fmt.Sprintf("%s/callback/doku", apiURL)
	}

	if customerInfo.RedirectURL != "" {
		chargeRequest.Order.CallbackURLResult = customerInfo.RedirectURL
		chargeRequest.Order.CallbackURLCancel = customerInfo.RedirectURL
	}

	// Set payment details
	chargeRequest.Payment.PaymentDueDate = 2880 // 48 hours
	// chargeRequest.Payment.Type = "SALE"
	chargeRequest.Payment.PaymentMethodTypes = []string{"CREDIT_CARD"}

	// Set customer details
	// chargeRequest.Customer.ID = customerInfo.UserId
	// chargeRequest.Customer.Name = customerInfo.CustomerName
	// if customerInfo.Email != "" {
	// 	chargeRequest.Customer.Email = customerInfo.Email
	// }
	// if customerInfo.PhoneNumber != "" {
	// 	chargeRequest.Customer.Phone = customerInfo.PhoneNumber
	// }
	// if customerInfo.Address != "" {
	// 	chargeRequest.Customer.Address = customerInfo.Address
	// }
	// if customerInfo.City != "" {
	// 	chargeRequest.Customer.City = customerInfo.City
	// }
	// if customerInfo.ProvinceState != "" {
	// 	chargeRequest.Customer.State = customerInfo.ProvinceState
	// }
	// if customerInfo.PostalCode != "" {
	// 	chargeRequest.Customer.Postcode = customerInfo.PostalCode
	// }
	// if customerInfo.Country != "" {
	// 	chargeRequest.Customer.Country = customerInfo.Country
	// }

	// Marshal struct to JSON
	jsonBody, err := json.Marshal(chargeRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	// Generate request headers
	requestID := uuid.New().String()
	requestTimestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	requestTarget := "/checkout/v1/payment"

	// Generate digest and signature
	digest := generateDigest(string(jsonBody))
	signature := generateRequestSignature(clientID, requestID, requestTimestamp, requestTarget, digest, secretKey)

	// Create HTTP request
	url := baseURL + requestTarget
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Request-Id", requestID)
	req.Header.Set("Request-Timestamp", requestTimestamp)
	req.Header.Set("Signature", signature)

	client := &http.Client{}
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var dokuResp DokuCheckoutResponse
	if err := json.Unmarshal(body, &dokuResp); err != nil {
		log.Printf("error unmarshal for transaction id: %s", transactionID)
		log.Println("response:", string(body))
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Log API call
	helper.DokuLogger.LogAPICall(
		url,
		"POST",
		time.Since(startTime),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionID,
			"request_body":   string(jsonBody),
			"headers": map[string]string{
				"Client-Id":         clientID,
				"Request-Id":        requestID,
				"Request-Timestamp": requestTimestamp,
				"Signature":         signature,
			},
		},
		map[string]interface{}{
			"body": string(body),
		},
	)

	now := time.Now()
	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("error charging for transaction id: %s, status: %d", transactionID, resp.StatusCode)
		return &dokuResp, fmt.Errorf("error response from DOKU: status %d", resp.StatusCode)
	}

	// Update transaction status to processing
	tokenID := dokuResp.Response.Payment.TokenID
	err = repository.UpdateTransactionStatus(context.Background(), transactionID, 1001, &tokenID, nil, "Processing payment", nil)
	if err != nil {
		log.Printf("Error updating transaction %s to PROCESSING: %s", transactionID, err)
	}

	return &dokuResp, nil
}
