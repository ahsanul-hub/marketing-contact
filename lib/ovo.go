package lib

import (
	"app/repository"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	ovoBatchCache = cache.New(5*time.Minute, 10*time.Minute)
	mu            sync.Mutex
)

func getRandomTransactionID() int {
	min := 500000
	max := 999999
	return rand.Intn(max-min+1) + min
}

func getIncrementedTransactionID() int {
	mu.Lock()
	defer mu.Unlock()

	counter, found := ovoBatchCache.Get("transactionID")
	if !found {

		counter = getRandomTransactionID()
	}

	if counter.(int) >= 999999 {
		counter = 500000
	} else {
		counter = counter.(int) + 1
	}

	ovoBatchCache.Set("transactionID", counter, cache.NoExpiration)
	return counter.(int)
}

type OVOTransactionRequest struct {
	Type                   string                 `json:"type"`
	ProcessingCode         string                 `json:"processingCode"`
	Amount                 uint                   `json:"amount"`
	Date                   string                 `json:"date"`
	ReferenceNumber        string                 `json:"referenceNumber"`
	Tid                    string                 `json:"tid"`
	Mid                    string                 `json:"mid"`
	MerchantId             string                 `json:"merchantId"`
	StoreCode              string                 `json:"storeCode"`
	AppSource              string                 `json:"appSource"`
	TransactionRequestData TransactionRequestData `json:"transactionRequestData"`
}

type TransactionRequestData struct {
	MerchantInvoice string `json:"merchantInvoice"`
	BatchNo         string `json:"batchNo"`
	Phone           string `json:"phone"`
}

type TransactionResponseData struct {
	StoreAddress1    string `json:"storeAddress1"`
	Ovoid            string `json:"ovoid"`
	CashUsed         string `json:"cashUsed"`
	StoreAddress2    string `json:"storeAddress2"`
	OvoPointsEarned  string `json:"ovoPointsEarned"`
	CashBalance      string `json:"cashBalance"`
	FullName         string `json:"fullName"`
	StoreName        string `json:"storeName"`
	OvoPointsUsed    string `json:"ovoPointsUsed"`
	OvoPointsBalance string `json:"ovoPointsBalance"`
	StoreCode        string `json:"storeCode"`
	PaymentType      string `json:"paymentType"`
}

type OVOResponse struct {
	Type                    string                   `json:"type"`
	ProcessingCode          string                   `json:"processingCode"`
	ApprovalCode            string                   `json:"approvalCode"`
	Amount                  int                      `json:"amount"`
	Date                    string                   `json:"date"`
	TraceNumber             int                      `json:"traceNumber"`
	ReferenceNumber         int                      `json:"referenceNumber"`
	HostTime                string                   `json:"hostTime"`
	MID                     string                   `json:"mid"`
	TID                     string                   `json:"tid"`
	HostDate                string                   `json:"hostDate"`
	ResponseCode            string                   `json:"responseCode"`
	TransactionRequestData  TransactionRequestData   `json:"transactionRequestData"`
	TransactionResponseData *TransactionResponseData `json:"transactionResponseData,omitempty"` // gunakan pointer dan `omitempty` agar tidak error saat kosong
}

func generateHMAC(appId, random, key string) string {
	data := appId + random
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ChargingOVO(transactionID string, amount uint, usermsisdn string) (*OVOResponse, error) {
	url := "https://api.ovo.id/pos"

	random := fmt.Sprintf("%d", time.Now().Unix())
	appId := "redision"
	appKey := "d7a0d2986d0998ca040a1eda66f390d2b9b636d5d0fe19e191e41c7e14a406ca"
	batchd := time.Now().Format("060102")

	nowDate := time.Now()

	batch := fmt.Sprintf("%06d", getIncrementedTransactionID())

	// invoice := fmt.Sprintf("%s%sRED-%s", invdate, milliseconds, transactionID)

	// Tanggal dalam format yang sesuai
	now := time.Now().Format("2006-01-02 15:04:05.000")

	requestBody := OVOTransactionRequest{
		Type:            "0200",
		ProcessingCode:  "040000",
		Amount:          amount,
		Date:            now,
		ReferenceNumber: batch,
		Tid:             "00020216",
		Mid:             "210638121215323",
		MerchantId:      "1810638",
		StoreCode:       "OLH2HPTPREDISIO",
		AppSource:       "POS",
		TransactionRequestData: TransactionRequestData{
			BatchNo:         batchd,
			MerchantInvoice: transactionID,
			Phone:           usermsisdn,
		},
	}

	// Marshal ke JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	// Buat HMAC

	hmacValue := generateHMAC(appId, random, appKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("app-id", "redision")
	req.Header.Set("hmac", hmacValue)
	req.Header.Set("random", random)
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
	defer cancel()

	req = req.WithContext(ctx)
	// Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var ovoResp OVOResponse
	if err := json.Unmarshal(bodyBytes, &ovoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	requestDate := &nowDate

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	err = repository.UpdateOvoRefBatch(context.Background(), transactionID, batchd, batch)
	if err != nil {
		log.Printf("Error updating ref number and batch number for transaction %s: %s", transactionID, err)
	}

	return &ovoResp, nil
}

func CheckStatusOVO(transactionID string, amount uint, usermsisdn, batchNo, referenceNumber string) (*OVOResponse, error) {
	url := "https://api.ovo.id/pos"

	random := fmt.Sprintf("%d", time.Now().Unix())
	appId := "redision"
	appKey := "d7a0d2986d0998ca040a1eda66f390d2b9b636d5d0fe19e191e41c7e14a406ca"

	nowDate := time.Now()

	// Tanggal dalam format yang sesuai
	now := time.Now().Format("2006-01-02 15:04:05.000")

	requestBody := OVOTransactionRequest{
		Type:            "0100",
		ProcessingCode:  "040000",
		Amount:          amount,
		Date:            now,
		ReferenceNumber: referenceNumber,
		Tid:             "00020216",
		Mid:             "210638121215323",
		MerchantId:      "1810638",
		StoreCode:       "OLH2HPTPREDISIO",
		AppSource:       "POS",
		TransactionRequestData: TransactionRequestData{
			BatchNo:         batchNo,
			MerchantInvoice: transactionID,
			Phone:           usermsisdn,
		},
	}

	// Marshal ke JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	// Buat HMAC

	hmacValue := generateHMAC(appId, random, appKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("app-id", "redision")
	req.Header.Set("hmac", hmacValue)
	req.Header.Set("random", random)
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 65*time.Second)
	defer cancel()

	req = req.WithContext(ctx)
	// Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var ovoResp OVOResponse
	if err := json.Unmarshal(bodyBytes, &ovoResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	requestDate := &nowDate

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionID, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionID, err)
	}

	return &ovoResp, nil
}
