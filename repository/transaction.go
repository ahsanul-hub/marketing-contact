package repository

import (
	"app/database"
	"app/dto/model"

	// "app/webhook"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

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

	transaction.AppID = client.ClientAppID
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

func stringToInt(value string) int {
	intVal, _ := strconv.Atoi(value)
	return intVal
}

func CreateTransaction(ctx context.Context, input *model.InputPaymentRequest, client *model.Client) (string, error) {
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
		Currency:      input.Currency,
		Price:         uint(chargingPrice),
		NetSettlement: float32(nettSettlement),
	}

	transaction.AppID = client.ClientAppID
	transaction.MerchantName = client.ClientName
	transaction.AppName = client.AppName
	transaction.ClientAppKey = client.ClientAppkey

	if err := database.DB.Create(&transaction).Error; err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction.ID, nil
}

func GetAllTransactions(ctx context.Context, limit, offset int, appID, userMDN, paymentMethod string, startDate, endDate *time.Time) ([]model.Transactions, int64, error) {
	span, _ := apm.StartSpan(ctx, "GetAllTransactions", "repository")
	defer span.End()
	var transactions []model.Transactions
	var totalItems int64

	query := database.DB

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if userMDN != "" {
		query = query.Where("user_mdn = ?", userMDN)
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}
	if err := query.Model(&model.Transactions{}).Where(query).Count(&totalItems).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count transactions: %w", err)
	}

	if err := query.Debug().Limit(limit).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	return transactions, totalItems, nil
}

func GetTransactionsMerchant(ctx context.Context, limit, offset int, appKey, appID, userMDN, paymentMethod string, startDate, endDate *time.Time) ([]model.TransactionMerchantResponse, int64, error) {
	var transactions []model.Transactions
	query := database.DB
	var totalItems int64

	if appKey != "" {
		query = query.Where("client_app_key = ?", appKey)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if userMDN != "" {
		query = query.Where("user_mdn = ?", userMDN)
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}
	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	if err := query.Model(&model.Transactions{}).Where(query).Count(&totalItems).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count transactions: %w", err)
	}

	if err := query.Select("user_mdn, user_id, payment_method, mt_tid AS merchant_transaction_id, status_code, amount, price, created_at, updated_at").
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
		UserMDN:                 transaction.UserMDN,
		UserID:                  transaction.UserId,
		PaymentMethod:           transaction.PaymentMethod,
		MerchantTransactionID:   transaction.MtTid,
		StatusCode:              transaction.StatusCode,
		TimestampRequestDate:    transaction.TimestampRequestDate,
		TimestampSubmitDate:     transaction.TimestampSubmitDate,
		TimestampCallbackDate:   transaction.TimestampCallbackDate,
		TimestampCallbackResult: transaction.TimestampCallbackResult,
		Route:                   transaction.Route,
		Amount:                  transaction.Amount,
		Price:                   transaction.Price,
		CreatedAt:               transaction.CreatedAt,
		UpdatedAt:               transaction.UpdatedAt,
	}

	return &response, nil
}

func UpdateTransactionStatusExpired(ctx context.Context, transactionID string, newStatusCode int, responseCallback string) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
	}
	timeLimit := time.Now().Add(-9 * time.Minute)

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? AND created_at <= ?", transactionID, timeLimit).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func UpdateTransactionStatus(ctx context.Context, transactionID string, newStatusCode int, responseCallback string) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
	}

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? ", transactionID).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func GetPendingTransactions(ctx context.Context) ([]model.Transactions, error) {
	var transactions []model.Transactions
	// timeLimit := time.Now().Add(-8 * time.Minute)

	if err := database.DB.Select("id, merchant_name", "status_code").Where("status_code = ?", 1001).Find(&transactions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return transactions, nil
		}
		return nil, fmt.Errorf("error fetching transactions: %w", err)
	}
	// log.Println("transactions: ", transactions)
	return transactions, nil
}

func UpdateTransactionCallbackTimestamps(ctx context.Context, transactionID string, callbackDate *time.Time, callbackResult string) error {
	db := database.DB

	updates := make(map[string]interface{})

	if callbackDate != nil {
		updates["timestamp_callback_date"] = callbackDate
	}
	if callbackResult != "" {
		updates["timestamp_callback_result"] = callbackResult
	}

	updates["status_code"] = 1000

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

		if db.RowsAffected == 0 {
			return fmt.Errorf("no transaction found with ID: %s", transactionID)
		}
	}

	return nil
}

func ProcessTransactions() {
	var transactions []model.Transactions

	if err := database.DB.Where("status_code = ?", 1003).Find(&transactions).Error; err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	for _, transaction := range transactions {
		arrClient, err := FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
		if err != nil {
			fmt.Println("Error fetching client:", err)
		}

		statusCode := 1000
		message := "Transaction updated"

		CallbackQueue <- CallbackJob{
			MerchantURL:   arrClient.CallbackURL,
			TransactionID: transaction.ID,
			MtTid:         transaction.MtTid,
			StatusCode:    statusCode,
			Message:       message,
		}
	}
}

type CallbackJob struct {
	MerchantURL   string `json:"merchant_url"`
	TransactionID string `json:"transaction_id"`
	MtTid         string `json:"merchant_transaction_id"`
	StatusCode    int    `json:"status_code"`
	Message       string `json:"message"`
}

var CallbackQueue = make(chan CallbackJob, 100)

func SendCallback(merchantURL string, transactionID string, mtTid string, statusCode int, message string) error {
	callbackData := CallbackJob{
		TransactionID: transactionID,
		StatusCode:    statusCode,
		MtTid:         mtTid,
		Message:       message,
	}

	jsonData, err := json.Marshal(callbackData)
	if err != nil {
		log.Printf("failed to marshal callback data: %v", err)
	}

	resp, err := http.Post(merchantURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to send callback: %v", err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		log.Printf("failed to decode response body: %v", err)
	}

	var callbackResult string
	if result, ok := responseBody["result"]; ok && result != nil {
		callbackResult = fmt.Sprintf("%v", result)
	} else {
		callbackResult = "ok" // Nilai default jika result nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("callback failed with status: %s", resp.Status)
	}

	// Update timestamp callback date dan result
	ctx := context.Background() // Buat context untuk digunakan dalam pembaruan
	callbackDate := time.Now()  // Ambil waktu sekarang sebagai callback date

	if err := UpdateTransactionCallbackTimestamps(ctx, transactionID, &callbackDate, callbackResult); err != nil {
		return fmt.Errorf("failed to update transaction callback timestamps: %v", err)
	}

	return nil
}

func sendCallbackWithRetry(merchantURL string, transactionID string, mtTid string, statusCode int, message string, retries int) {
	for i := 0; i < retries; i++ {
		err := SendCallback(merchantURL, transactionID, mtTid, statusCode, message)
		if err == nil {
			fmt.Println("Callback sent successfully")
			return
		}

		fmt.Printf("Failed to send callback, attempt %d: %v\n", i+1, err)
		time.Sleep(5 * time.Minute)
	}

	fmt.Println("All retry attempts failed")
}

func ProcessCallbackQueue() {
	for job := range CallbackQueue {
		sendCallbackWithRetry(job.MerchantURL, job.TransactionID, job.MtTid, job.StatusCode, job.Message, 5)
	}
}
