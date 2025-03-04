package lib

import (
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

type GopayChargeRequest struct {
	PaymentType        string `json:"payment_type"`
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount uint   `json:"gross_amount"`
	} `json:"transaction_details"`
	Gopay struct {
		CallbackURL string `json:"callback_url"`
	} `json:"gopay"`
}

func RequestChargingGopay(transactionID string, chargingPrice uint) (*MidtransResponse, error) {
	chargeRequest := GopayChargeRequest{
		PaymentType: "gopay",
	}
	chargeRequest.TransactionDetails.OrderID = transactionID

	chargeRequest.TransactionDetails.GrossAmount = chargingPrice
	chargeRequest.Gopay.CallbackURL = "https://new-payment.redision.com/api/callback/midtrans"

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

	var midtransResp MidtransResponse
	if err := json.Unmarshal(body, &midtransResp); err != nil {
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
		return &midtransResp, fmt.Errorf("error response from Midtrans: %s", midtransResp.StatusMessage)
	}

	return &midtransResp, nil
}
