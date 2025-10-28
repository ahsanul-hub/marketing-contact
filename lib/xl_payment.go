package lib

import (
	"app/config"
	"app/dto/model"
	"app/helper"
	"app/repository"
	"app/worker"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/patrickmn/go-cache"
)

var tokenCache = cache.New(56*time.Minute, 58*time.Minute)
var NumberCache = cache.New(2*time.Minute, 3*time.Minute)

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type CustomerInfo struct {
	MSISDN           string `json:"msisdn"`
	SubscriberNo     string `json:"subscriberNo"`
	SubscriberType   string `json:"subscriberType"`
	SubscriberStatus string `json:"subscriberStatus"`
}

type InquiryStatus struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}

type CheckNumberResponse struct {
	CustomerInfo  CustomerInfo  `json:"CustomerInfo"`
	InquiryStatus InquiryStatus `json:"InquiryStatus"`
}

type RequestChargingInfo struct {
	UserIdentifier     string `json:"userIdentifier"`
	UserIdentifierType string `json:"userIdentifierType"`
}

type TransactionStatus struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}

type TransactionInfo struct {
	TransactionID   string `json:"transactionId"`
	PartnerID       string `json:"partnerId"`
	Item            string `json:"item"`
	ItemDescription string `json:"itemDescription"`
	BalanceType     string `json:"balanceType"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	RefferenceId    string `json:"refferenceId"`
}

type ChargingResponse struct {
	CustomerInfo      RequestChargingInfo `json:"CustomerInfo"`
	TransactionStatus TransactionStatus   `json:"TransactionStatus"`
	TransactionInfo   TransactionInfo     `json:"TransactionInfo"`
}

// Struct utama untuk permintaan charging
type ChargingRequest struct {
	CustomerInfo    RequestChargingInfo `json:"customerInfo"`
	TransactionInfo TransactionInfo     `json:"transactionInfo"`
}
type TransactionInquiryResponse struct {
	CustomerInfo      CustomerInfo      `json:"CustomerInfo"`
	TransactionInfo   TransactionInfo   `json:"TransactionInfo"`
	TransactionStatus TransactionStatus `json:"TransactionStatus"`
}

type TransactionInquiryStatusResponse struct {
	CustomerInfo      interface{}              `json:"transactionInquiryCustomerInfoTO"`
	TransactionInfo   TransactionInfoStatus    `json:"transactionInquiryInfoTO"`
	TransactionStatus TransactionInquiryStatus `json:"transactionInquiryStatusTO"`
}

type CustomerInfoStatus struct {
	UserIdentifier string `json:"userIdentifier"`
}
type TransactionInfoStatus struct {
	TransactionId string  `json:"transactionId"`
	ReferenceId   *string `json:"referenceId"`
}
type TransactionInquiryStatus struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}

func RequestToken(clientID, clientSecret string) (*TokenResponse, error) {
	config, _ := config.GetGatewayConfig("xl_twt")
	arrayOptions := config.Options["development"].(map[string]interface{})

	requestBody := TokenRequest{
		ClientID:     arrayOptions["clientid"].(string),
		ClientSecret: arrayOptions["clientsecret"].(string),
		GrantType:    "client_credentials",
	}
	url := arrayOptions["tokenurl"].(string)
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

func CheckNumberXl(msisdn string, token string) (bool, error) {
	config, _ := config.GetGatewayConfig("xl_twt")
	arrayOptions := config.Options["development"].(map[string]interface{})

	baseURL := arrayOptions["inquiryurl"].(string)

	url := fmt.Sprintf("%s?msisdn=%s&partnerId=RDSN", baseURL, msisdn)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("cache-control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Decode the response body
	var checkNumberResponse CheckNumberResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkNumberResponse); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}
	// Check response code and return the appropriate boolean
	if checkNumberResponse.InquiryStatus.ResponseCode == "00" || checkNumberResponse.InquiryStatus.ResponseCode == "21" {
		return true, nil
	} else if checkNumberResponse.InquiryStatus.ResponseCode == "20" {
		NumberCache.Set(msisdn, "xl_airtime", cache.DefaultExpiration)
		return false, fmt.Errorf(checkNumberResponse.InquiryStatus.ResponseDesc)
	}

	return false, fmt.Errorf("unexpected response code: %s", checkNumberResponse.InquiryStatus.ResponseCode)
}

func GetAccessTokenXl(clientID, clientSecret string) (string, error) {

	if cachedToken, found := tokenCache.Get("accessToken"); found {
		// log.Println("Token diambil dari cache.")
		return cachedToken.(string), nil
	}

	tokenResp, err := RequestToken(clientID, clientSecret)
	if err != nil {
		return "", err
	}

	tokenCache.Set("accessToken", tokenResp.AccessToken, cache.DefaultExpiration)
	// log.Println("Token baru diminta dan disimpan ke cache.")

	return tokenResp.AccessToken, nil
}

func RequestChargingXL(msisdn, itemID, itemDesc, transactionId string, chargingPrice uint) (ChargingResponse, error) {
	now := time.Now()
	config, _ := config.GetGatewayConfig("xl_twt")
	arrayOptions := config.Options["development"].(map[string]interface{})
	url := arrayOptions["chargingurl"].(string)

	token, err := GetAccessTokenXl(arrayOptions["clientid"].(string), arrayOptions["clientsecret"].(string))
	if err != nil {
		return ChargingResponse{}, err
	}
	beautifyMsisdn := helper.BeautifyIDNumber(msisdn, false)

	isNumberActive, err := CheckNumberXl(beautifyMsisdn, token)
	if !isNumberActive {
		log.Println("err:", err)
		err := repository.UpdateTransactionStatus(context.Background(), transactionId, 1005, nil, nil, "Msisdn not active", nil)
		if err != nil {
			log.Println("err: ", err)
		}
		return ChargingResponse{}, fmt.Errorf("invalid phone number or not found")
	}

	chargingRequest := ChargingRequest{
		CustomerInfo: RequestChargingInfo{
			UserIdentifier:     beautifyMsisdn,
			UserIdentifierType: "MSISDN",
		},
		TransactionInfo: TransactionInfo{
			TransactionID:   transactionId,
			PartnerID:       "RDSN",
			Item:            itemID,
			ItemDescription: itemDesc,
			BalanceType:     "AirBalance",
			Amount:          fmt.Sprintf("%d", chargingPrice),
			Currency:        "IDR",
		},
	}

	body, err := json.Marshal(chargingRequest)
	if err != nil {
		return ChargingResponse{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return ChargingResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("cache-control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ChargingResponse{}, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChargingResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	helper.XLLogger.LogAPICall(
		url,
		"POST",
		time.Since(now),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionId,
			"request_body":   string(body),
		},
		map[string]interface{}{
			"body": string(bodyBytes),
		},
	)

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return ChargingResponse{}, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	// Decode the response body
	var responseMap struct {
		ChargingResponse ChargingResponse `json:"chargingResponse"`
	}

	if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
		return ChargingResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return responseMap.ChargingResponse, nil
}

func CheckTransactions(transactionID, partnerID, token string) (TransactionInquiryStatusResponse, error) {
	config, err := config.GetGatewayConfig("xl_twt")
	if err != nil {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("failed to get gateway config: %w", err)
	}

	arrayOptions := config.Options["development"].(map[string]interface{})
	baseURL := arrayOptions["checkurl"].(string)
	url := fmt.Sprintf("%s?transactionId=%s&partnerId=%s", baseURL, transactionID, partnerID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("cache-control", "no-cache")

	dumpReq, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Printf("Failed to dump request: %v\n", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var responseMap struct {
		TransactionInquiryResponse TransactionInquiryStatusResponse `json:"TransactionInquiryResponse"`
	}

	// Decode the response body using the updated structure
	if err := json.Unmarshal(bodyBytes, &responseMap); err != nil {
		return TransactionInquiryStatusResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if responseMap.TransactionInquiryResponse.TransactionStatus.ResponseCode == "00" {
		helper.XLLogger.LogCallback(transactionID, true,
			map[string]interface{}{
				"transaction_id":   transactionID,
				"method":           "GET",
				"request_callback": string(dumpReq),
				"response":         string(bodyBytes),
			},
		)
	}

	// Mengembalikan status response inquiry jika ada
	return responseMap.TransactionInquiryResponse, nil
}

func CheckTransactionStatus(transaction model.Transactions) {
	// Ambil token
	config, _ := config.GetGatewayConfig("xl_twt")
	arrayOptions := config.Options["development"].(map[string]interface{})

	token, err := GetAccessTokenXl(arrayOptions["clientid"].(string), arrayOptions["clientsecret"].(string))
	if err != nil {
		log.Printf("Error getting access token for transaction %s: %s", transaction.ID, err)
		return
	}

	// Periksa status transaction
	response, err := CheckTransactions(transaction.ID, "RDSN", token)
	if err != nil {
		log.Printf("Error checking transaction %s: %s", transaction.ID, err)
		return
	}

	referenceId := response.TransactionInfo.ReferenceId

	now := time.Now()

	receiveCallbackDate := &now

	if response.TransactionStatus.ResponseCode == "00" {

		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, referenceId, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
		}
		log.Printf("%s, transaction ID : %s, responseCode: %s", response.TransactionStatus.ResponseDesc, transaction.ID, response.TransactionStatus.ResponseCode)
	} else {
		createdAt := transaction.CreatedAt
		timeLimit := time.Now().Add(-10 * time.Minute)

		expired := createdAt.Before(timeLimit)
		if expired {
			if err := repository.UpdateTransactionStatusExpired(context.Background(), transaction.ID, 1005, "", ""); err != nil {
				log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
			}
		}

	}
}

func ProcessPendingTransactions() {

	for {
		go worker.ProcessTransactions()
		go worker.ProcessFailedTransactions()
		// transactions, err := repository.GetPendingTransactions(context.Background(), "xl_airtime")

		// if err != nil {
		// 	log.Printf("Error retrieving pending transactions: %s", err)
		// 	time.Sleep(1 * time.Minute)
		// 	continue
		// }
		// for _, transaction := range transactions {
		// 	go CheckTransactionStatus(transaction)
		// }

		time.Sleep(5 * time.Second)
	}
}
