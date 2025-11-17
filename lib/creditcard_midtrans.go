package lib

import (
	"app/config"
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type CreditCardChargeRequest struct {
	PaymentType        string `json:"payment_type"`
	TransactionDetails struct {
		OrderID     string `json:"order_id"`
		GrossAmount uint   `json:"gross_amount"`
	} `json:"transaction_details"`
	CreditCard struct {
		TokenID        string `json:"token_id"`
		Authentication bool   `json:"authentication"`
	} `json:"credit_card"`
	ItemDetails struct {
		Price    uint   `json:"price"`
		Quantity int    `json:"quantity"`
		Name     string `json:"name"`
	} `json:"item_details"`
	CustomerDetails struct {
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Email     string `json:"email,omitempty"`
		Phone     string `json:"phone,omitempty"`
	} `json:"customer_details,omitempty"`
	Callbacks struct {
		Finish string `json:"finish"`
	} `json:"callbacks,omitempty"`
}

// RequestChargingCreditCard performs Midtrans Core API credit card charge, expecting 3DS with redirect_url
func RequestChargingCreditCard(transactionID string, amount uint, tokenID string, finishCallbackURL string, customerName string, email string, phone string, itemName string) (*MidtransResponse, error) {
	reqBody := CreditCardChargeRequest{
		PaymentType: "credit_card",
	}
	reqBody.TransactionDetails.OrderID = transactionID
	reqBody.TransactionDetails.GrossAmount = amount
	reqBody.CreditCard.TokenID = tokenID
	reqBody.CreditCard.Authentication = true
	if finishCallbackURL != "" {
		reqBody.Callbacks.Finish = finishCallbackURL
	}

	reqBody.ItemDetails.Name = itemName
	reqBody.ItemDetails.Quantity = 1
	reqBody.ItemDetails.Price = amount

	if strings.TrimSpace(customerName) != "" {
		nameParts := strings.Fields(customerName)
		reqBody.CustomerDetails.FirstName = nameParts[0]
		if len(nameParts) > 1 {
			reqBody.CustomerDetails.LastName = strings.Join(nameParts[1:], " ")
		}
	}
	if email != "" {
		reqBody.CustomerDetails.Email = email
	}
	if phone != "" {
		reqBody.CustomerDetails.Phone = phone
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	log.Println("json body credit card cc: ", string(jsonBody))

	baseURL := config.Config("MIDTRANS_BASE_URL", "https://api.sandbox.midtrans.com")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/charge", strings.TrimRight(baseURL, "/")), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	serverKey := config.Config("MIDTRANS_SERVER_KEY", "")
	if serverKey == "" {
		return nil, fmt.Errorf("midtrans server key is not configured")
	}
	authToken := base64.StdEncoding.EncodeToString([]byte(serverKey + ":"))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authToken))

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

	helper.MidtransLogger.LogAPICall(
		fmt.Sprintf("%s/v2/charge", strings.TrimRight(baseURL, "/")),
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
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	now := time.Now()
	requestDate := &now
	if err := repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil); err != nil {
		// continue even if timestamp update fails
	}

	// Midtrans returns 201 for charge created; for CC with 3DS expect redirect_url
	if midtransResp.StatusCode != "201" {
		return &midtransResp, fmt.Errorf("error response from Midtrans: %s", midtransResp.StatusMessage)
	}

	return &midtransResp, nil
}
