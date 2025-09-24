package lib

import (
	"app/config"
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type CreditCardChargingRequest struct {
	ClientReferenceId string `json:"clientReferenceId"`
	Amount            struct {
		Value    int    `json:"value"`
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
	Customer struct {
		GivenName   string `json:"givenName"`
		Email       string `json:"email"`
		PhoneNumber struct {
			CountryCode string `json:"countryCode"`
			Number      string `json:"number"`
		} `json:"phoneNumber"`
	} `json:"customer"`
	OrderInformation struct {
		ProductDetails []ProductDetail `json:"productDetails"`
	} `json:"orderInformation"`
	AutoConfirm         bool   `json:"autoConfirm"`
	StatementDescriptor string `json:"statementDescriptor"`
	ExpiryAt            string `json:"expiryAt"`
}

type ProductDetail struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
	Price    Price  `json:"price"`
}

type Price struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}

func CardHarsyaCharging(transactionId, customerName, userMdn string, amount uint) (*HarsyaChargingResponse, error) {
	clientId := config.Config("PIVOT_CLIENT_ID", "")
	clientSecret := config.Config("PIVOT_CLIENT_SECRET", "")
	accessToken, err := GetAccessTokenHarsya(clientId, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timeNow := time.Now().In(loc)
	expiryAt := timeNow.Add(60 * time.Minute)

	successUrl := fmt.Sprintf("%s/return/dana", config.Config("APIURL", ""))
	requestUrl := fmt.Sprintf("%s/v2/payments", config.Config("PIVOT_BASE_URL", ""))

	requestBody := CreditCardChargingRequest{
		ClientReferenceId:   transactionId,
		Mode:                "REDIRECT",
		ExpiryAt:            expiryAt.Format(time.RFC3339),
		AutoConfirm:         true,
		StatementDescriptor: "Redision",
	}

	requestBody.Amount.Value = int(amount)
	requestBody.Amount.Currency = "IDR"
	requestBody.PaymentMethod.Type = "CARD"

	requestBody.RedirectUrl.SuccessReturnUrl = successUrl
	requestBody.RedirectUrl.FailureReturnUrl = "https://merchant.com/failure"
	requestBody.RedirectUrl.ExpirationReturnUrl = "https://merchant.com/expiration"

	requestBody.Customer.GivenName = customerName
	requestBody.Customer.Email = "customeremail@example.com"
	requestBody.Customer.PhoneNumber.CountryCode = "+62"

	requestBody.OrderInformation.ProductDetails = []ProductDetail{
		{
			Type:     "DIGITAL",
			Quantity: 1,
			Price: Price{
				Value:    int(amount),
				Currency: "IDR",
			},
		},
	}

	phoneNumber := strings.TrimPrefix(userMdn, "0")
	requestBody.Customer.PhoneNumber.Number = phoneNumber

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(body))
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

	var chargingResp HarsyaChargingResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	helper.HarsyaLogger.LogAPICall(
		requestUrl,
		"POST",
		time.Since(now),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionId,
			"request_body":   requestBody,
		},
		map[string]interface{}{
			"body": chargingResp,
		},
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
		} else {
			log.Printf("Error response from Harsya API (Status: %s): %s", resp.Status, string(responseBody))
		}
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

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
