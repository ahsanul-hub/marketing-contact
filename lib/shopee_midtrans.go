package lib

import (
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ShopeePayChargeRequest struct {
	PaymentType        string `json:"payment_type"`
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount uint   `json:"gross_amount"`
	} `json:"transaction_details"`
	ShopeePay struct {
		CallbackURL string `json:"callback_url"`
	} `json:"shopeepay"`
}

type MidtransResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id,omitempty"`
	OrderID           string `json:"order_id,omitempty"`
	MerchantID        string `json:"merchant_id,omitempty"`
	GrossAmount       string `json:"gross_amount,omitempty"`
	Currency          string `json:"currency,omitempty"`
	PaymentType       string `json:"payment_type,omitempty"`
	TransactionTime   string `json:"transaction_time,omitempty"`
	TransactionStatus string `json:"transaction_status,omitempty"`
	FraudStatus       string `json:"fraud_status,omitempty"`
	Actions           []struct {
		Name   string `json:"name,omitempty"`
		Method string `json:"method,omitempty"`
		URL    string `json:"url,omitempty"`
	} `json:"actions,omitempty"`
	QrString               string `json:"qr_string,omitempty"`
	ChannelResponseCode    string `json:"channel_response_code,omitempty"`
	ChannelResponseMessage string `json:"channel_response_message,omitempty"`
	ExpiryTime             string `json:"expiry_time,omitempty"`
	ID                     string `json:"id,omitempty"`
}

func RequestChargingShopeePay(transactionID string, chargingPrice uint, redirectUrl string) (*MidtransResponse, error) {
	chargeRequest := ShopeePayChargeRequest{
		PaymentType: "shopeepay",
	}
	chargeRequest.TransactionDetails.OrderID = transactionID

	chargeRequest.TransactionDetails.GrossAmount = chargingPrice

	var backUrl string

	if redirectUrl != "" {
		backUrl = redirectUrl
	} else {
		backUrl = "https://new-payment.redision.com/api/callback/midtrans"
	}
	chargeRequest.ShopeePay.CallbackURL = backUrl

	// Marshal struct menjadi JSON
	jsonBody, err := json.Marshal(chargeRequest)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	// Membuat HTTP request
	req, err := http.NewRequest("POST", "https://api.midtrans.com/v2/charge", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic TWlkLXNlcnZlci1MU2puUUNiMW0zcDhsSzEyVm9mbF9tZF86")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	start := time.Now()
	helper.ShopeepayLogger.LogAPICall(
		"https://api.midtrans.com/v2/charge",
		"POST",
		time.Since(start),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionID,
			"request_body":   jsonBody,
		},
		map[string]interface{}{
			"body": string(body),
		},
	)

	var midtransResp MidtransResponse
	if err := json.Unmarshal(body, &midtransResp); err != nil {
		log.Printf("error unmarshal for transaction id :%s", transactionID)
		log.Println("res", string(body))
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	now := time.Now()

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	if midtransResp.StatusCode != "201" {
		log.Printf("error charging for transaction id :%s message : %s", transactionID, midtransResp.StatusMessage)
		return &midtransResp, fmt.Errorf("error response from Midtrans: %s", midtransResp.StatusMessage)
	}

	return &midtransResp, nil
}
