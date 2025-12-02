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
		// BillingInfo    BillingInfo     `json:"billingInfo,omitempty"`
		// ShippingInfo   ShippingInfo    `json:"shippingInfo,omitempty"`
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

type BillingInfo struct {
	GivenName     string      `json:"givenName"`
	SureName      string      `json:"sureName"`
	Email         string      `json:"email"`
	PhoneNumber   PhoneNumber `json:"phoneNumber"`
	AddressLine1  string      `json:"addressLine1"`
	AddressLine2  string      `json:"addressLine2"`
	City          string      `json:"city"`
	ProvinceState string      `json:"provinceState"`
	Country       string      `json:"country"`
	PostalCode    string      `json:"postalCode"`
}

type ShippingInfo struct {
	GivenName     string      `json:"givenName"`
	SureName      string      `json:"sureName"`
	Email         string      `json:"email"`
	PhoneNumber   PhoneNumber `json:"phoneNumber"`
	AddressLine1  string      `json:"addressLine1"`
	AddressLine2  string      `json:"addressLine2"`
	City          string      `json:"city"`
	ProvinceState string      `json:"provinceState"`
	Country       string      `json:"country"`
	PostalCode    string      `json:"postalCode"`
	Method        string      `json:"method"`
	ShippingFee   struct {
		Value    int    `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
}
type PhoneNumber struct {
	CountryCode string `json:"countryCode"`
	Number      string `json:"number"`
}

type Price struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}

// HarsyaPaymentSessionResponse untuk response create payment session dengan encryption key
type HarsyaPaymentSessionResponse struct {
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Data    HarsyaPaymentSessionData `json:"data"`
}

type HarsyaPaymentSessionData struct {
	ID                string `json:"id"`
	ClientReferenceID string `json:"clientReferenceId"`
	EncryptionKey     string `json:"encryptionKey"`
	Amount            struct {
		Currency string `json:"currency"`
		Value    int    `json:"value"`
	} `json:"amount"`
	Mode          string      `json:"mode"`
	Status        string      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
	ExpiryAt      string      `json:"expiryAt"`
	RedirectURL   RedirectURL `json:"redirectUrl"`
	PaymentMethod struct {
		Type string `json:"type"`
	} `json:"paymentMethod"`
}

// HarsyaConfirmRequest untuk confirm payment session dengan encrypted card
type HarsyaConfirmRequest struct {
	PaymentMethod struct {
		Type string `json:"type"`
		Card struct {
			EncryptedCard string `json:"encryptedCard"`
		} `json:"card"`
	} `json:"paymentMethod"`
	PaymentMethodOptions struct {
		Card struct {
			CaptureMethod string `json:"captureMethod"`
			ThreeDsMethod string `json:"threeDsMethod"`
		} `json:"card"`
	} `json:"paymentMethodOptions"`
}

// CreateHarsyaPaymentSession membuat payment session dengan mode API untuk mendapatkan encryption key
func CreateHarsyaPaymentSession(clientRefId, transactionId, customerName, userMdn, redirectUrl, email, address, provinceState, country, postalCode, city, countryCode, phoneNumber string, amount uint) (*HarsyaPaymentSessionResponse, error) {
	clientId := config.Config("PIVOT_CLIENT_ID", "")
	clientSecret := config.Config("PIVOT_CLIENT_SECRET", "")
	accessToken, err := GetAccessTokenHarsya(clientId, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	loc := time.FixedZone("IST", 5*60*60+30*60)
	now := time.Now().In(loc)
	timeNow := time.Now().In(loc)
	expiryAt := timeNow.Add(24 * time.Hour)

	baseURL := config.Config("APIURL", "")
	successUrl := fmt.Sprintf("%s/payment/success?transaction_id=%s", baseURL, clientRefId)

	// Use custom redirect URL if provided
	if redirectUrl != "" {
		successUrl = redirectUrl
	}

	requestUrl := fmt.Sprintf("%s/v2/payments", config.Config("PIVOT_BASE_URL", ""))

	requestBody := CreditCardChargingRequest{
		ClientReferenceId:   clientRefId,
		Mode:                "API", // Mode API untuk mendapatkan encryption key
		ExpiryAt:            expiryAt.Format(time.RFC3339),
		AutoConfirm:         false, // false untuk mendapatkan encryption key
		StatementDescriptor: "Redision",
	}

	requestBody.Amount.Value = int(amount)
	requestBody.Amount.Currency = "IDR"
	requestBody.PaymentMethod.Type = "CARD"

	failureUrl := fmt.Sprintf("%s/payment/failure?transaction_id=%s", baseURL, clientRefId)
	expirationUrl := fmt.Sprintf("%s/payment/expiration?transaction_id=%s", baseURL, clientRefId)

	requestBody.RedirectUrl.SuccessReturnUrl = successUrl
	requestBody.RedirectUrl.FailureReturnUrl = failureUrl
	requestBody.RedirectUrl.ExpirationReturnUrl = expirationUrl

	requestBody.Customer.GivenName = customerName
	if email != "" {
		requestBody.Customer.Email = email
	} else {
		requestBody.Customer.Email = "customeremail@example.com"
	}
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

	phoneNumber = strings.TrimPrefix(userMdn, "0")
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

	log.Printf("Sending Harsya Payment Session Request: %+v", requestBody)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var logBody interface{}
	var sessionResp HarsyaPaymentSessionResponse

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		if err := json.Unmarshal(respBodyBytes, &sessionResp); err == nil {
			logBody = sessionResp
		} else {
			logBody = string(respBodyBytes)
		}
	} else {
		logBody = string(respBodyBytes)
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
			"body": logBody,
		},
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("Error response from Harsya API (Status: %s): %s", resp.Status, string(respBodyBytes))
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	requestDate := &now
	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	err = repository.UpdateTransactionStatus(context.Background(), transactionId, 1001, &sessionResp.Data.ID, nil, "Processing payment", nil)
	if err != nil {
		log.Printf("Error updating transaction %s to PROCESSING: %s", transactionId, err)
	}

	return &sessionResp, nil
}

// ConfirmHarsyaPaymentSession mengkonfirmasi payment session dengan encrypted card
func ConfirmHarsyaPaymentSession(paymentSessionId, encryptedCard string) (*HarsyaChargingResponse, error) {
	clientId := config.Config("PIVOT_CLIENT_ID", "")
	clientSecret := config.Config("PIVOT_CLIENT_SECRET", "")
	accessToken, err := GetAccessTokenHarsya(clientId, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	now := time.Now()
	requestUrl := fmt.Sprintf("%s/v2/payments/%s/confirm", config.Config("PIVOT_BASE_URL", ""), paymentSessionId)

	requestBody := HarsyaConfirmRequest{
		PaymentMethod: struct {
			Type string `json:"type"`
			Card struct {
				EncryptedCard string `json:"encryptedCard"`
			} `json:"card"`
		}{
			Type: "CARD",
			Card: struct {
				EncryptedCard string `json:"encryptedCard"`
			}{
				EncryptedCard: encryptedCard,
			},
		},
		PaymentMethodOptions: struct {
			Card struct {
				CaptureMethod string `json:"captureMethod"`
				ThreeDsMethod string `json:"threeDsMethod"`
			} `json:"card"`
		}{
			Card: struct {
				CaptureMethod string `json:"captureMethod"`
				ThreeDsMethod string `json:"threeDsMethod"`
			}{
				CaptureMethod: "automatic",
				ThreeDsMethod: "CHALLENGE",
			},
		},
	}

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

	var confirmResp HarsyaChargingResponse
	if err := json.NewDecoder(resp.Body).Decode(&confirmResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	helper.HarsyaLogger.LogAPICall(
		requestUrl,
		"POST",
		time.Since(now),
		resp.StatusCode,
		map[string]interface{}{
			"payment_session_id": paymentSessionId,
			"request_body":       requestBody,
		},
		map[string]interface{}{
			"body": confirmResp,
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

	return &confirmResp, nil
}

func CardHarsyaCharging(transactionId, customerName, userMdn, redirectUrl, email, address, provinceState, country, postalCode, city, countryCode, phoneNumber string, amount uint) (*HarsyaChargingResponse, error) {
	clientId := config.Config("PIVOT_CLIENT_ID", "")
	clientSecret := config.Config("PIVOT_CLIENT_SECRET", "")
	accessToken, err := GetAccessTokenHarsya(clientId, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	now := time.Now()
	loc, _ := time.LoadLocation("Asia/Jakarta")
	timeNow := time.Now().In(loc)
	expiryAt := timeNow.Add(24 * time.Hour)

	successUrl := fmt.Sprintf("%s/return/dana", config.Config("APIURL", ""))

	if redirectUrl != "" {
		successUrl = redirectUrl
	}

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
	// requestBody.OrderInformation.BillingInfo = BillingInfo{
	// 	GivenName: customerName,
	// 	SureName:  "",
	// 	Email:     email,
	// 	PhoneNumber: PhoneNumber{
	// 		CountryCode: countryCode,
	// 		Number:      phoneNumber,
	// 	},
	// 	AddressLine1:  address,
	// 	AddressLine2:  "",
	// 	City:          city,
	// 	ProvinceState: provinceState,
	// 	Country:       country,
	// 	PostalCode:    postalCode,
	// }

	phoneNumber = strings.TrimPrefix(userMdn, "0")
	requestBody.Customer.PhoneNumber.Number = phoneNumber

	// requestBody.OrderInformation.ShippingInfo = ShippingInfo{
	// 	GivenName: customerName,
	// 	SureName:  "",
	// 	Email:     email,
	// 	PhoneNumber: PhoneNumber{
	// 		CountryCode: countryCode,
	// 		Number:      phoneNumber,
	// 	},
	// 	AddressLine1:  address,
	// 	AddressLine2:  "",
	// 	City:          city,
	// 	ProvinceState: provinceState,
	// 	Country:       country,
	// 	PostalCode:    postalCode,
	// 	Method:        "REGULAR",
	// }
	// requestBody.OrderInformation.ShippingInfo.ShippingFee.Value = 0
	// requestBody.OrderInformation.ShippingInfo.ShippingFee.Currency = "IDR"

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
