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

type QrisChargeRequest struct {
	PaymentType        string `json:"payment_type"`
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount uint   `json:"gross_amount"`
	} `json:"transaction_details"`
	Qris struct {
		Acquirer string `json:"acquirer"`
	} `json:"qris"`
}

func RequestChargingQris(transactionID string, chargingPrice uint) (*MidtransResponse, error) {
	chargeRequest := QrisChargeRequest{
		PaymentType: "qris",
	}
	chargeRequest.TransactionDetails.OrderID = transactionID

	chargeRequest.TransactionDetails.GrossAmount = chargingPrice
	chargeRequest.Qris.Acquirer = "gopay"

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
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	helper.QrisLogger.LogAPICall(
		"https://api.midtrans.com/v2/charge",
		"POST",
		time.Since(start),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionID,
			"request_body":   string(jsonBody),
		},
		map[string]interface{}{
			"body": string(body),
		},
	)

	var midtransResp MidtransResponse
	if err := json.Unmarshal(body, &midtransResp); err != nil {
		log.Println("res", string(body))
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	now := time.Now()

	// log.Println("test main log")

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	if midtransResp.StatusCode != "201" {
		return &midtransResp, fmt.Errorf("error response from Midtrans: %s", midtransResp.StatusMessage)
	}

	return &midtransResp, nil
}
