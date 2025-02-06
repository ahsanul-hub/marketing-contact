package repository

import (
	"app/database"
	"app/dto/model"
	"crypto/hmac"
	"crypto/sha256"
	"strings"
	"sync"

	// "app/webhook"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var processedTransactions sync.Map

func CheckTransaction(transactionID, appKey, appID string) (*model.Transactions, error) {
	ctx := context.Background()

	collection := database.GetCollection("dcb", "transactions")

	filter := bson.M{
		"merchant_transaction_id": transactionID,
		"client_appkey":           appKey,
		"app_id":                  appID,
	}

	var result model.Transactions

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

func CreateOrder(ctx context.Context, input *model.InputPaymentRequest, client *model.Client) (uint, error) {

	uniqueID, err := uuid.NewV7()
	if err != nil {
		log.Fatal(err)
	}

	transaction := model.Transactions{
		ID:            uniqueID.String(),
		ClientAppKey:  input.ClientAppKey,
		StatusCode:    1001,
		ItemName:      input.ItemName,
		UserMDN:       input.UserMDN,
		Testing:       input.Testing,
		Route:         input.Route,
		PaymentMethod: input.PaymentMethod,
		Currency:      input.Currency,
		Price:         input.Price,
	}

	transaction.AppID = client.ClientID
	transaction.MerchantName = client.ClientName
	transaction.ClientAppKey = client.ClientAppkey

	collection := database.GetCollection("dcb", "transactions")
	result, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return 0, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return 0, fmt.Errorf("could not convert inserted ID to ObjectID")
	}

	return uint(id.Timestamp().Unix()), nil
}

// func stringToInt(value string) int {
// 	intVal, _ := strconv.Atoi(value)
// 	return intVal
// }

func CreateTransaction(ctx context.Context, input *model.InputPaymentRequest, client *model.Client) (string, uint, error) {
	span, _ := apm.StartSpan(ctx, "CreateTransaction", "repository")
	defer span.End()
	uniqueID, err := uuid.NewV7()
	if err != nil {
		log.Fatal(err)
	}
	settlementConfig, err := GetSettlementConfig(client.UID)
	if err != nil {
		log.Println(err)
	}

	var selectedSettlement *model.SettlementClient
	for _, settlement := range settlementConfig {

		if settlement.Name == input.PaymentMethod {
			selectedSettlement = &settlement
			break
		}
	}

	additionalPercent := 0.11
	if selectedSettlement.AdditionalPercent != nil {
		additionalPercent = float64(*selectedSettlement.AdditionalPercent) / 100
	}

	chargingPrice := float64(input.Amount)*additionalPercent + float64(input.Amount)
	nettSettlement := float64(input.Amount) * (float64(*selectedSettlement.SharePartner) / 100)

	currency := input.Currency
	if currency == "" {
		currency = "IDR"
	}

	transaction := model.Transactions{
		ID:            uniqueID.String(),
		ClientAppKey:  input.ClientAppKey,
		MtTid:         input.MtTid,
		StatusCode:    1001,
		ItemName:      input.ItemName,
		UserMDN:       input.UserMDN,
		Testing:       input.Testing,
		Route:         input.Route,
		UserId:        input.UserId,
		PaymentMethod: input.PaymentMethod,
		Currency:      currency,
		Price:         uint(chargingPrice),
		NetSettlement: float32(nettSettlement),
		Amount:        input.Amount,
		ItemId:        input.ItemId,
		BodySign:      input.BodySign,
	}

	transaction.AppID = client.ClientID
	transaction.MerchantName = client.ClientName
	transaction.AppName = client.AppName
	transaction.ClientAppKey = client.ClientAppkey

	if err := database.DB.Create(&transaction).Error; err != nil {
		return "", 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction.ID, transaction.Price, nil
}

func GetAllTransactions(ctx context.Context, limit, offset, status, denom int, transactionId, merchantTransactionId, appID, userMDN, userId, appName string, merchants, paymentMethods []string, startDate, endDate *time.Time) ([]model.Transactions, int64, error) {
	span, _ := apm.StartSpan(ctx, "GetAllTransactions", "repository")
	defer span.End()
	var transactions []model.Transactions
	var totalItems int64

	query := database.DB

	if transactionId != "" {
		query = query.Where("id = ?", transactionId)
	}
	if merchantTransactionId != "" {
		query = query.Where("mt_tid = ?", merchantTransactionId)
	}
	if status != 0 {
		query = query.Where("status_code = ?", status)
	}
	if denom != 0 {
		query = query.Where("amount = ?", denom)
	}
	if userId != "" {
		query = query.Where("user_id = ?", userId)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}
	if len(merchants) > 0 {
		query = query.Where("merchant_name IN ?", merchants)
	}
	if userMDN != "" {
		query = query.Where("user_mdn = ?", userMDN)
	}
	if len(paymentMethods) > 0 {
		query = query.Where("payment_method IN ?", paymentMethods)
	}
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}
	if err := query.Model(&model.Transactions{}).Where(query).Count(&totalItems).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count transactions: %w", err)
	}

	if err := query.Debug().Order("created_at DESC").Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	return transactions, totalItems, nil
}

func GetTransactionsMerchant(ctx context.Context, limit, offset, status, denom int, merchantTransactionId, appKey, appID, userMDN, appName string, paymentMethods []string, startDate, endDate *time.Time) ([]model.TransactionMerchantResponse, int64, error) {
	var transactions []model.Transactions
	query := database.DB
	var totalItems int64

	if merchantTransactionId != "" {
		query = query.Where("mt_tid = ?", merchantTransactionId)
	}
	if appKey != "" {
		query = query.Where("client_app_key = ?", appKey)
	}
	if status != 0 {
		query = query.Where("status_code = ?", status)
	}
	if denom != 0 {
		query = query.Where("amount = ?", denom)
	}
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if userMDN != "" {
		query = query.Where("user_mdn = ?", userMDN)
	}
	if len(paymentMethods) > 0 {
		query = query.Where("payment_method IN ?", paymentMethods)
	}
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	if err := query.Model(&model.Transactions{}).Where(query).Count(&totalItems).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count transactions: %w", err)
	}

	if err := query.Select("user_mdn, user_id, payment_method, mt_tid , status_code, amount, price, item_name, item_id,app_name, created_at, updated_at").Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	var response []model.TransactionMerchantResponse
	for _, transaction := range transactions {
		response = append(response, model.TransactionMerchantResponse{
			UserMDN:                 transaction.UserMDN,
			UserID:                  transaction.UserId,
			PaymentMethod:           transaction.PaymentMethod,
			MerchantTransactionID:   transaction.MtTid,
			StatusCode:              transaction.StatusCode,
			TimestampRequestDate:    transaction.TimestampRequestDate,
			TimestampSubmitDate:     transaction.TimestampSubmitDate,
			TimestampCallbackDate:   transaction.TimestampCallbackDate,
			TimestampCallbackResult: transaction.TimestampCallbackResult,
			ItemName:                transaction.ItemName,
			ItemId:                  transaction.ItemId,
			Currency:                transaction.Currency,
			AppName:                 transaction.AppName,
			Route:                   transaction.Route,
			Amount:                  transaction.Amount,
			Price:                   transaction.Price,
			CreatedAt:               transaction.CreatedAt,
			UpdatedAt:               transaction.UpdatedAt,
		})
	}

	return response, totalItems, nil
}

func GetTransactionByID(ctx context.Context, id string) (*model.Transactions, error) {
	var transaction model.Transactions
	if err := database.DB.Where("id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}
	return &transaction, nil
}

func GetTransactionMerchantByID(ctx context.Context, appKey, appId, id string) (*model.TransactionMerchantResponse, error) {
	var transaction model.Transactions
	if err := database.DB.Where("mt_tid = ? AND client_app_key = ? AND app_id = ?", id, appKey, appId).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}

	response := model.TransactionMerchantResponse{
		ID:                      transaction.ID,
		UserMDN:                 transaction.UserMDN,
		UserID:                  transaction.UserId,
		PaymentMethod:           transaction.PaymentMethod,
		MerchantTransactionID:   transaction.MtTid,
		StatusCode:              transaction.StatusCode,
		TimestampRequestDate:    transaction.TimestampRequestDate,
		TimestampSubmitDate:     transaction.TimestampSubmitDate,
		TimestampCallbackDate:   transaction.TimestampCallbackDate,
		TimestampCallbackResult: transaction.TimestampCallbackResult,
		ItemName:                transaction.ItemName,
		ItemId:                  transaction.ItemId,
		Currency:                transaction.Currency,
		AppName:                 transaction.AppName,
		Route:                   transaction.Route,
		Amount:                  transaction.Amount,
		Price:                   transaction.Price,
		CreatedAt:               transaction.CreatedAt,
		UpdatedAt:               transaction.UpdatedAt,
	}

	return &response, nil
}

func UpdateTransactionStatusExpired(ctx context.Context, transactionID string, newStatusCode int, responseCallback, failReason string) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
		FailReason: failReason,
	}
	timeLimit := time.Now().Add(-9 * time.Minute)

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? AND created_at <= ?", transactionID, timeLimit).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

// func UpdateTransactionFailReason(ctx context.Context, transactionID string, FailReason string) error {
// 	db := database.DB

// 	transactionUpdate := model.Transactions{
// 		FailReason: FailReason,
// 	}

// 	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? ", transactionID).Updates(transactionUpdate).Error; err != nil {
// 		return fmt.Errorf("failed to update fail reason: %w", err)
// 	}

// 	return nil
// }

// func UpdateTransactionStatusFailCallback(ctx context.Context, transactionID string, newStatusCode int, responseCallback string) error {
// 	db := database.DB

// 	transactionUpdate := model.Transactions{
// 		StatusCode: newStatusCode,
// 	}

// 	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? ", transactionID).Updates(transactionUpdate).Error; err != nil {
// 		return fmt.Errorf("failed to update transaction status: %w", err)
// 	}

// 	return nil
// }

func UpdateTransactionStatus(ctx context.Context, transactionID string, newStatusCode int, ximpayId *string, failReason string) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
	}

	if ximpayId != nil {
		transactionUpdate.XimpayID = *ximpayId
	}

	if failReason != "" {
		transactionUpdate.FailReason = failReason
	}

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? AND status_code = ", transactionID, 1001).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func GetPendingTransactions(ctx context.Context, paymentMethod string) ([]model.Transactions, error) {
	var transactions []model.Transactions
	// timeLimit := time.Now().Add(-8 * time.Minute)

	// if err := database.DB.Select("id, merchant_name", "status_code").Where("status_code = ?", 1001).Find(&transactions).Error; err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		return transactions, nil
	// 	}
	// 	return nil, fmt.Errorf("error fetching transactions: %w", err)
	// }

	query := database.DB.Select("id, merchant_name, status_code").Where("status_code = ?", 1001)

	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("error fetching pending transactions: %w", err)
	}

	return transactions, nil
}

func UpdateTransactionCallbackTimestamps(ctx context.Context, transactionID string, statusCode int, callbackDate *time.Time, callbackResult string) error {
	db := database.DB

	updates := make(map[string]interface{})

	if callbackDate != nil {
		updates["timestamp_callback_date"] = callbackDate
	}
	if callbackResult != "" {
		updates["timestamp_callback_result"] = callbackResult
	}

	updates["status_code"] = statusCode

	if len(updates) > 0 {
		if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ?", transactionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update transaction callback timestamps: %w", err)
		}

	}

	return nil
}

func UpdateTransactionTimestamps(ctx context.Context, transactionID string, requestDate, submitDate, callbackDate *time.Time) error {
	db := database.DB

	updates := make(map[string]interface{})

	if requestDate != nil {
		updates["timestamp_request_date"] = requestDate
	}
	if submitDate != nil {
		updates["timestamp_submit_date"] = submitDate
	}
	if callbackDate != nil {
		updates["timestamp_callback_date"] = callbackDate
	}

	if len(updates) > 0 {
		if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ?", transactionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update transaction timestamps: %w", err)
		}
	}

	return nil
}

// func ProcessTransactions() {
// 	var transactions []model.Transactions

// 	if err := database.DB.Where("status_code = ?", 1003).Find(&transactions).Error; err != nil {
// 		fmt.Println("Error fetching transactions:", err)
// 		return
// 	}

// 	for _, transaction := range transactions {

// 		arrClient, err := FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
// 		if err != nil {
// 			fmt.Println("Error fetching client:", err)
// 		}

// 		statusCode := 1000

// 		callbackData := CallbackData{
// 			UserID:                transaction.UserId,
// 			MerchantTransactionID: transaction.MtTid,
// 			StatusCode:            statusCode,
// 			PaymentMethod:         transaction.PaymentMethod,
// 			Amount:                fmt.Sprintf("%d", transaction.Amount),
// 			Status:                "success",
// 			Currency:              transaction.Currency,
// 			ItemName:              transaction.ItemName,
// 			ItemID:                transaction.ItemId,
// 			ReferenceID:           transaction.ReferenceID,
// 		}

// 		CallbackQueue <- CallbackQueueStruct{
// 			Data:          callbackData,
// 			TransactionId: transaction.ID,
// 			Secret:        arrClient.ClientSecret,
// 			MerchantURL:   arrClient.CallbackURL,
// 		}
// 		time.Sleep(10 * time.Second)

// 	}
// 	log.Println("transactions length:", len(transactions))
// }

func ProcessTransactions() {

	var transactions []model.Transactions
	// if err := database.DB.Where("status_code = ? AND timestamp_callback_result != ?", 1003, "failed").Find(&transactions).Error; err != nil {
	// 	fmt.Println("Error fetching transactions:", err)
	// 	return
	// }
	err := database.DB.Raw("SELECT id, mt_tid, payment_method, amount, client_app_key, app_id, currency,item_name, item_id, reference_id, status_code FROM transactions WHERE status_code = ? AND timestamp_callback_result != ?", 1003, "failed").Scan(&transactions).Error
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	for _, transaction := range transactions {
		// Cek apakah transaksi sudah diproses
		if _, loaded := processedTransactions.LoadOrStore(transaction.ID, true); loaded {
			// Jika sudah diproses, lewati transaksi ini
			continue
		}
		// Proses transaksi dalam goroutine
		go func(transaction model.Transactions) {
			arrClient, err := FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
			if err != nil {
				log.Printf("Error fetching client for transaction %s: %v", transaction.ID, err)
				return
			}
			// Siapkan data callback
			callbackData := CallbackData{
				UserID:                transaction.UserId,
				MerchantTransactionID: transaction.MtTid,
				StatusCode:            1000, // Misalnya, status sukses
				PaymentMethod:         transaction.PaymentMethod,
				Amount:                fmt.Sprintf("%d", transaction.Amount),
				Status:                "success",
				Currency:              transaction.Currency,
				ItemName:              transaction.ItemName,
				ItemID:                transaction.ItemId,
				ReferenceID:           transaction.ReferenceID,
			}

			// Kirim ke CallbackQueue
			CallbackQueue <- CallbackQueueStruct{
				Data:          callbackData,
				TransactionId: transaction.ID,
				Secret:        arrClient.ClientSecret,
				MerchantURL:   arrClient.CallbackURL,
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
	UserID                string `json:"user_id"`
	MerchantTransactionID string `json:"merchant_transaction_id"`
	StatusCode            int    `json:"status_code"`
	PaymentMethod         string `json:"payment_method"`
	Amount                string `json:"amount"`
	Status                string `json:"status"`
	Currency              string `json:"currency"`
	ItemName              string `json:"item_name"`
	ItemID                string `json:"item_id"`
	ReferenceID           string `json:"reference_id"`
}
type CallbackQueueStruct struct {
	Data          CallbackData
	TransactionId string
	Secret        string
	MerchantURL   string
}

var CallbackQueue = make(chan CallbackQueueStruct, 100)

func SendCallback(merchantURL, secret string, transactionID string, data CallbackData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to marshal callback data: %v", err)
	}

	bodyJSONString := string(jsonData)
	// log.Println("jsonData", bodyJSONString)

	bodySign, _ := GenerateBodySign(bodyJSONString, secret)
	// log.Println("bodySign", bodySign)

	req, err := http.NewRequest(http.MethodPost, merchantURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return err
	}

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
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		// log.Printf("failed to decode response body: %v", err)
	}

	var callbackResult string
	if result, ok := responseBody["result"]; ok && result != nil {
		callbackResult = fmt.Sprintf("%v", result)
	} else {
		callbackResult = "ok"
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback failed with status: %s , bodySign: %s", resp.Status, bodySign)
	}

	// Update timestamp callback date dan result
	ctx := context.Background() // Buat context untuk digunakan dalam pembaruan
	callbackDate := time.Now()  // Ambil waktu sekarang sebagai callback date

	if err := UpdateTransactionCallbackTimestamps(ctx, transactionID, 1000, &callbackDate, callbackResult); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return nil
}

func sendCallbackWithRetry(merchantURL string, transactionID string, secret string, retries int, data CallbackData) error {
	for i := 0; i < retries; i++ {

		err := SendCallback(merchantURL, secret, transactionID, data)
		if err == nil {
			fmt.Println("Callback sent successfully")
			return nil
		}

		time.Sleep(5 * time.Minute)
	}

	if err := UpdateTransactionCallbackTimestamps(context.Background(), transactionID, 1003, nil, "failed"); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return fmt.Errorf("all retry attempts failed for transactionId: %s", transactionID) // Kembalikan error jika semua percobaan gagal
}

// func ProcessCallbackQueue() {
// 	for job := range CallbackQueue {
// 		log.Printf("Processing callback for transactionId: %s", job.TransactionId)
// 		sendCallbackWithRetry(job.MerchantURL, job.TransactionId, job.Secret, 5, job.Data)
// 	}
// }

func ProcessCallbackQueue() {
	for job := range CallbackQueue {
		// Jalankan pengiriman callback dalam goroutine
		go func(job CallbackQueueStruct) {
			// log.Printf("Processing callback for transactionId: %s", job.TransactionId)
			err := sendCallbackWithRetry(job.MerchantURL, job.TransactionId, job.Secret, 5, job.Data)
			if err != nil {
				fmt.Printf("Failed to send callback for transactionId: %s: %v", job.TransactionId, err)
			}
		}(job)
	}
}

func GenerateBodySign(bodyJson string, appSecret string) (string, error) {

	h := hmac.New(sha256.New, []byte(appSecret))

	// Write the data (bodyJson) to the HMAC
	h.Write([]byte(bodyJson))

	// Get the HMAC result
	signature := h.Sum(nil)

	// Encode the HMAC result to Base64
	base64Encoded := base64.StdEncoding.EncodeToString(signature)

	bodysign := strings.NewReplacer("+", "-", "/", "_").Replace(base64Encoded)

	return bodysign, nil
}
