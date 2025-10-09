package worker

import (
	"app/database"
	"app/dto/model"
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CallbackQueueStruct struct {
	Data          interface{}
	TransactionId string
	Secret        string
	MerchantURL   string
}

var SuccessCallbackQueue = make(chan CallbackQueueStruct, 100)
var FailedCallbackQueue = make(chan CallbackQueueStruct, 100)
var processedTransactions sync.Map

func ProcessCallbackQueue() {
	for job := range SuccessCallbackQueue {
		// Jalankan pengiriman callback dalam goroutine
		go func(job CallbackQueueStruct) {
			// log.Printf("Processing callback for transactionId: %s", job.TransactionId)
			err := SendCallbackWithRetry(job.MerchantURL, job.TransactionId, job.Secret, 5, job.Data)
			if err != nil {
				fmt.Printf("Failed to send callback for transactionId: %s: %v", job.TransactionId, err)
			}
		}(job)
	}
}

func ProccessFailedCallbackWorker() {
	for job := range FailedCallbackQueue {
		go func(j CallbackQueueStruct) {
			err := SendCallbackFailedRetry(j.MerchantURL, j.TransactionId, j.Secret, 5, j.Data)
			if err != nil {
				log.Println("Callback failed, failed to send:", err)
			}
		}(job)
	}
}

func ProcessTransactions() {

	var transactions []model.Transactions

	err := database.DB.Raw("SELECT id, mt_tid, payment_method, amount, client_app_key, app_id, currency, item_name, item_id, user_id, reference_id, ximpay_id, midtrans_transaction_id, status_code, notification_url, callback_reference_id FROM transactions WHERE status_code = ? AND timestamp_callback_result != ?", 1003, "failed").Scan(&transactions).Error
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	for _, transaction := range transactions {
		if _, loaded := processedTransactions.LoadOrStore(transaction.ID, true); loaded {
			continue
		}
		// Proses transaksi dalam goroutine
		go func(transaction model.Transactions) {
			arrClient, err := repository.FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
			var callbackURL string
			for _, app := range arrClient.ClientApps {
				if app.AppID == transaction.AppID {
					callbackURL = app.CallbackURL
					break
				}
			}

			if callbackURL == "" {
				log.Printf("No matching ClientApp found for AppID: %s", transaction.AppID)
				return
			}

			if transaction.NotificationUrl != "" {
				callbackURL = transaction.NotificationUrl
			}

			if err != nil {
				log.Printf("Error fetching client for transaction %s: %v", transaction.ID, err)
				return
			}

			var paymentMethod string

			paymentMethod = transaction.PaymentMethod
			if arrClient.ClientName == "HIGO GAME PTE LTD" && transaction.PaymentMethod == "qris" {
				paymentMethod = "qr"
			}

			var amount interface{}
			if arrClient.ClientName == "LeisureLink Digital Limited" || arrClient.ClientSecret == "o_G0JIzzJLditvj" {
				amount = transaction.Amount
			} else {
				amount = fmt.Sprintf("%d", transaction.Amount)
			}

			var callbackPayload interface{}

			if arrClient.ClientName == "PM Max" || arrClient.ClientSecret == "gmtb50vcf5qcvwr" ||
				arrClient.ClientName == "Coda" || arrClient.ClientSecret == "71mczdtiyfaunj5" ||
				arrClient.ClientName == "TutuReels" || arrClient.ClientSecret == "UPF6qN7b2nP5geg" ||
				arrClient.ClientName == "Redigame2" || arrClient.ClientSecret == "gjq7ygxhztmlkgg" {
				callbackPayload = model.CallbackDataLegacy{
					AppID:                  transaction.AppID,
					ClientAppKey:           transaction.ClientAppKey,
					UserID:                 transaction.UserId,
					UserIP:                 transaction.UserIP,
					UserMDN:                transaction.UserMDN,
					MerchantTransactionID:  transaction.MtTid,
					TransactionDescription: "",
					PaymentMethod:          paymentMethod,
					Currency:               transaction.Currency,
					Amount:                 transaction.Amount,
					ChargingAmount:         fmt.Sprintf("%d", transaction.Price),
					StatusCode:             "1000",
					Status:                 "success",
					ItemID:                 transaction.ItemId,
					ItemName:               transaction.ItemName,
					UpdatedAt:              fmt.Sprintf("%d", time.Now().Unix()),
					ReferenceID:            transaction.CallbackReferenceId,
					Testing:                "0",
					Custom:                 "",
				}
			} else {
				payload := CallbackData{
					UserID:                transaction.UserId,
					MerchantTransactionID: transaction.MtTid,
					StatusCode:            1000,
					PaymentMethod:         paymentMethod,
					Amount:                amount,
					Status:                "success",
					Currency:              transaction.Currency,
					ItemName:              transaction.ItemName,
					ItemID:                transaction.ItemId,
					ReferenceID:           transaction.ID,
				}

				if arrClient.ClientName == "Zingplay International PTE,. LTD" || arrClient.ClientSecret == "9qyxr81YWU2BNlO" {
					payload.AppID = transaction.AppID
					payload.ClientAppKey = transaction.ClientAppKey
				}

				callbackPayload = payload
			}

			SuccessCallbackQueue <- CallbackQueueStruct{
				Data:          callbackPayload,
				TransactionId: transaction.ID,
				Secret:        arrClient.ClientSecret,
				MerchantURL:   callbackURL,
			}
		}(transaction)
	}
}

func ProcessFailedTransactions() {

	var transactions []model.Transactions

	err := database.DB.Raw(`
	SELECT 
		t.id, t.mt_tid, t.payment_method, t.amount, t.client_app_key, t.app_id, 
		t.currency, t.item_name, t.item_id, t.user_id, t.reference_id, t.callback_reference_id,
		t.ximpay_id, t.midtrans_transaction_id, t.status_code 
	FROM 
		transactions t
	INNER JOIN 
		client_apps ca 
		ON t.client_app_key = ca.app_key AND t.app_id = ca.app_id
	WHERE 
		t.status_code = ? 
		AND (t.timestamp_callback_result IS NULL OR t.timestamp_callback_result = '')  
		AND t.created_at >= NOW() - INTERVAL '1 days'
		AND ca.fail_callback = '1'
`, 1005).Scan(&transactions).Error
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	for _, transaction := range transactions {
		if _, loaded := processedTransactions.LoadOrStore(transaction.ID, true); loaded {
			continue
		}

		go func(transaction model.Transactions) {
			arrClient, err := repository.FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
			if err != nil {
				log.Printf("Error fetching client for transaction %s: %v", transaction.ID, err)
				return
			}

			if arrClient == nil || len(arrClient.ClientApps) == 0 {
				log.Printf("No client data found for transaction %s (AppKey: %s, AppID: %s)", transaction.ID, transaction.ClientAppKey, transaction.AppID)
				return
			}

			if arrClient.FailCallback == "0" {
				return
			}

			var callbackURL string
			for _, app := range arrClient.ClientApps {
				if app.AppID == transaction.AppID {
					callbackURL = app.CallbackURL
					break
				}
			}

			if callbackURL == "" {
				log.Printf("No matching ClientApp found for AppID: %s", transaction.AppID)
				return
			}

			var paymentMethod string
			var status string

			paymentMethod = transaction.PaymentMethod
			if transaction.MerchantName == "HIGO GAME PTE LTD" && transaction.PaymentMethod == "qris" {
				paymentMethod = "qr"
			}

			switch transaction.StatusCode {
			case 1005:
				status = "failed"
			case 1001:
				status = "pending"
			}

			var amount interface{}
			if arrClient.ClientName == "LeisureLink Digital Limited" || arrClient.ClientSecret == "o_G0JIzzJLditvj" {
				amount = transaction.Amount
			} else {
				amount = fmt.Sprintf("%d", transaction.Amount)
			}

			// callbackData := CallbackData{
			// 	UserID:                transaction.UserId,
			// 	MerchantTransactionID: transaction.MtTid,
			// 	StatusCode:            transaction.StatusCode,
			// 	PaymentMethod:         paymentMethod,
			// 	Amount:                amount,
			// 	Status:                status,
			// 	Currency:              transaction.Currency,
			// 	ItemName:              transaction.ItemName,
			// 	ItemID:                transaction.ItemId,
			// 	ReferenceID:           transaction.ID,
			// }
			// if arrClient.ClientName == "Zingplay International PTE,. LTD" || arrClient.ClientSecret == "9qyxr81YWU2BNlO" {
			// 	callbackData.AppID = transaction.AppID
			// 	callbackData.ClientAppKey = transaction.ClientAppKey
			// }

			var callbackPayload interface{}

			if arrClient.ClientName == "PM Max" || arrClient.ClientSecret == "gmtb50vcf5qcvwr" ||
				arrClient.ClientName == "Coda" || arrClient.ClientSecret == "71mczdtiyfaunj5" ||
				arrClient.ClientName == "TutuReels" || arrClient.ClientSecret == "UPF6qN7b2nP5geg" ||
				arrClient.ClientName == "Redigame2" || arrClient.ClientSecret == "gjq7ygxhztmlkgg" {
				callbackPayload = model.FailedCallbackDataLegacy{
					AppID:                  transaction.AppID,
					ClientAppKey:           transaction.ClientAppKey,
					UserID:                 transaction.UserId,
					UserIP:                 transaction.UserIP,
					UserMDN:                transaction.UserMDN,
					MerchantTransactionID:  transaction.MtTid,
					TransactionDescription: "",
					PaymentMethod:          paymentMethod,
					Currency:               transaction.Currency,
					Amount:                 transaction.Amount,
					StatusCode:             fmt.Sprintf("%d", transaction.StatusCode),
					Status:                 status,
					ItemID:                 transaction.ItemId,
					ItemName:               transaction.ItemName,
					UpdatedAt:              fmt.Sprintf("%d", time.Now().Unix()),
					ReferenceID:            transaction.CallbackReferenceId,
					Testing:                "0",
					Custom:                 "",
					FailReason:             status,
				}
			} else {
				payload := CallbackData{
					UserID:                transaction.UserId,
					MerchantTransactionID: transaction.MtTid,
					StatusCode:            transaction.StatusCode,
					PaymentMethod:         paymentMethod,
					Amount:                amount,
					Status:                status,
					Currency:              transaction.Currency,
					ItemName:              transaction.ItemName,
					ItemID:                transaction.ItemId,
					ReferenceID:           transaction.ID,
				}

				if arrClient.ClientName == "Zingplay International PTE,. LTD" || arrClient.ClientSecret == "9qyxr81YWU2BNlO" {
					payload.AppID = transaction.AppID
					payload.ClientAppKey = transaction.ClientAppKey
				}

				callbackPayload = payload
			}

			FailedCallbackQueue <- CallbackQueueStruct{
				Data:          callbackPayload,
				TransactionId: transaction.ID,
				Secret:        arrClient.ClientSecret,
				MerchantURL:   callbackURL,
			}
		}(transaction)
	}
}

type CallbackJob struct {
	MerchantURL   string `json:"merchant_url"`
	TransactionID string `json:"transaction_id"`
	MtTid         string `json:"merchant_transaction_id"`
	StatusCode    int    `json:"status_code"`
	Message       string `json:"message"`
}

type CallbackData struct {
	UserID                string      `json:"user_id"`
	MerchantTransactionID string      `json:"merchant_transaction_id"`
	StatusCode            int         `json:"status_code"`
	PaymentMethod         string      `json:"payment_method"`
	Amount                interface{} `json:"amount"`
	Status                string      `json:"status"`
	Currency              string      `json:"currency"`
	ItemName              string      `json:"item_name"`
	ItemID                string      `json:"item_id"`
	ReferenceID           string      `json:"reference_id"`
	AppID                 string      `json:"app_id,omitempty"`
	ClientAppKey          string      `json:"client_appkey,omitempty"`
}

// CallbackLogger interface untuk menghindari import cycle
type CallbackLogger interface {
	LogAPICall(endpoint, method string, duration time.Duration, statusCode int, requestData, responseData map[string]interface{})
}

func SendCallback(merchantURL, secret string, transactionID string, data interface{}) (map[string]interface{}, error) {
	return SendCallbackWithLogger(merchantURL, secret, transactionID, data, nil)
}

func SendCallbackWithLogger(merchantURL, secret string, transactionID string, data interface{}, logger CallbackLogger) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal callback data: %v", err)
		return nil, err
	}

	bodyJSONString := string(jsonData)
	// log.Println("jsonData Callback", bodyJSONString)

	bodySign, _ := repository.GenerateBodySign(bodyJSONString, secret)
	// log.Println("bodySign", bodySign)

	req, err := http.NewRequest(http.MethodPost, merchantURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("bodysign", bodySign)

	client := &http.Client{}
	callbackDate := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send callback: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	// Read entire body to handle non-JSON responses safely
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("failed to read response body: %v", readErr)
	} else {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "application/json") {
			if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
				log.Printf("failed to decode response body as JSON: %v", err)
				responseBody = map[string]interface{}{"raw_body": string(bodyBytes)}
			}
		} else {
			// Non-JSON response (e.g., text/html, text/plain)
			responseBody = map[string]interface{}{
				"raw_body":     string(bodyBytes),
				"content_type": contentType,
			}
		}
	}

	var callbackResult string
	if result, ok := responseBody["result"]; ok && result != nil {
		callbackResult = fmt.Sprintf("%v", result)
	} else {
		callbackResult = "ok"
	}

	if logger != nil {
		logger.LogAPICall(
			merchantURL,
			"POST",
			time.Since(callbackDate),
			resp.StatusCode,
			map[string]interface{}{
				"transaction_id":  transactionID,
				"type":            "callback success",
				"header_bodysign": bodySign,
				"request_body":    bodyJSONString,
			},
			map[string]interface{}{
				"body": responseBody,
			},
		)
	} else {
		log.Printf("Callback response for transaction %s: status=%d, body=%+v", transactionID, resp.StatusCode, responseBody)
	}

	if resp.StatusCode != http.StatusOK {
		return responseBody, fmt.Errorf("callback failed with status: %s , url: %s", resp.Status, merchantURL)
	}

	ctx := context.Background()

	if err := repository.UpdateTransactionCallbackTimestamps(ctx, transactionID, 1000, &callbackDate, callbackResult); err != nil {
		return responseBody, fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return responseBody, nil
}

func SendCallbackFailed(merchantURL, secret string, transactionID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal callback data: %v", err)
	}

	bodyJSONString := string(jsonData)
	// log.Println("jsonData Callback Failed", bodyJSONString)
	callbackDate := time.Now()

	bodySign, _ := repository.GenerateBodySign(bodyJSONString, secret)

	req, err := http.NewRequest(http.MethodPost, merchantURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return err
	}

	// log.Println("callback failed data send:", bodyJSONString)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("bodysign", bodySign)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send callback: %v", err)
		return err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("failed to read response body: %v", readErr)
	} else {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(contentType), "application/json") {
			if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
				log.Printf("failed to decode response body as JSON: %v", err)
				responseBody = map[string]interface{}{"raw_body": string(bodyBytes)}
			}
		} else {
			responseBody = map[string]interface{}{
				"raw_body":     string(bodyBytes),
				"content_type": contentType,
			}
		}
	}

	var callbackResult string
	if result, ok := responseBody["result"]; ok && result != nil {
		callbackResult = fmt.Sprintf("%v", result)
	} else {
		callbackResult = "ok"
	}

	helper.NotificationLogger.LogAPICall(
		merchantURL,
		"POST",
		time.Since(callbackDate),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id":  transactionID,
			"type":            "callback failed",
			"header_bodysign": bodySign,
			"request_body":    bodyJSONString,
		},
		map[string]interface{}{
			"body": responseBody,
		},
	)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback failed with status: %s , bodySign: %s, transactionId: %s", resp.Status, bodySign, transactionID)
	}

	ctx := context.Background()

	var statusCode int
	switch v := data.(type) {
	case CallbackData:
		statusCode = v.StatusCode
	case model.FailedCallbackDataLegacy:
		code, err := strconv.Atoi(v.StatusCode)
		if err != nil {
			return fmt.Errorf("invalid status code in legacy callback data: %v", err)
		}
		statusCode = code
	default:
		return fmt.Errorf("unsupported callback data type: %T", data)
	}

	if err := repository.UpdateTransactionCallbackTimestamps(ctx, transactionID, statusCode, &callbackDate, callbackResult); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return nil
}

func SendCallbackWithRetry(merchantURL string, transactionID string, secret string, retries int, data interface{}) error {
	for i := 0; i < retries; i++ {

		_, err := SendCallbackWithLogger(merchantURL, secret, transactionID, data, helper.NotificationLogger)

		if err == nil {
			fmt.Println("Callback sent successfully")
			return nil
		}

		time.Sleep(5 * time.Minute)
	}

	if err := repository.UpdateTransactionCallbackTimestamps(context.Background(), transactionID, 1003, nil, "failed"); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return fmt.Errorf("all retry attempts failed for transactionId: %s", transactionID)
}

func SendCallbackFailedRetry(merchantURL string, transactionID string, secret string, retries int, data interface{}) error {
	for i := 0; i < retries; i++ {

		err := SendCallbackFailed(merchantURL, secret, transactionID, data)
		if err == nil {
			fmt.Println("Callback failed sent successfully")
			return nil
		}

		time.Sleep(5 * time.Minute)
	}

	var statusCode int
	switch v := data.(type) {
	case CallbackData:
		statusCode = v.StatusCode
	case model.FailedCallbackDataLegacy:
		code, err := strconv.Atoi(v.StatusCode)
		if err != nil {
			return fmt.Errorf("invalid status code in legacy callback data: %v", err)
		}
		statusCode = code
	default:
		return fmt.Errorf("unsupported callback data type: %T", data)
	}

	if err := repository.UpdateTransactionCallbackTimestamps(context.Background(), transactionID, statusCode, nil, "failed"); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return fmt.Errorf("all retry attempts failed for transactionId: %s", transactionID) // Kembalikan error jika semua percobaan gagal
}
