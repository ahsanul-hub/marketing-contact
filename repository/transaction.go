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
		StatusCode:    1,
		ItemName:      input.ItemName,
		UserMDN:       input.Mobile,
		Testing:       input.Testing,
		Route:         input.Route,
		PaymentMethod: input.PaymentMethod,
		Currency:      input.Currency,
		Price:         input.Price,
	}

	transaction.AppID = client.ClientAppID
	transaction.MerchantName = client.ClientName
	transaction.AppKey = client.ClientAppkey

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
		StatusCode:    1,
		ItemName:      input.ItemName,
		UserMDN:       input.Mobile,
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
	transaction.AppKey = client.ClientAppkey

	if err := database.DB.Create(&transaction).Error; err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction.ID, nil
}

func GetAllTransactions(ctx context.Context) ([]model.Transactions, error) {
	var transactions []model.Transactions
	if err := database.DB.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch transactions: %w", err)
	}
	return transactions, nil
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

func UpdateTransactionStatus(ctx context.Context, transactionID string, newStatusCode int, responseCallback string) error {
	// Dapatkan koneksi ke database
	db := database.DB

	// Buat struktur untuk menyimpan perubahan status
	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
	}

	// Lakukan pembaruan ke database
	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ?", transactionID).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Cek apakah ada rekaman yang diperbarui
	if db.RowsAffected == 0 {
		return fmt.Errorf("no transaction found with ID: %s", transactionID)
	}

	return nil
}

func GetPendingTransactions(ctx context.Context) ([]model.Transactions, error) {
	var transactions []model.Transactions

	if err := database.DB.Where("status_code = ?", 1001).Find(&transactions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return transactions, nil
		}
		return nil, fmt.Errorf("error fetching transactions: %w", err)
	}
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

	if len(updates) > 0 {
		if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ?", transactionID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update transaction callback timestamps: %w", err)
		}

		if db.RowsAffected == 0 {
			return fmt.Errorf("no transaction found with ID: %s", transactionID)
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

	if err := database.DB.Where("status_code = ?", 1000).Find(&transactions).Error; err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	for _, transaction := range transactions {
		arrClient, err := FindClient(transaction.AppKey, transaction.AppID)
		if err != nil {
			fmt.Println("Error fetching client:", err)
		}

		statusCode := 1000
		message := "Transaction updated"

		CallbackQueue <- CallbackJob{
			MerchantURL:   arrClient.CallbackURL,
			TransactionID: transaction.ID,
			StatusCode:    statusCode,
			Message:       message,
		}
	}
}

type CallbackJob struct {
	MerchantURL   string
	TransactionID string
	StatusCode    int
	Message       string
}

var CallbackQueue = make(chan CallbackJob, 100) // Antrean dengan kapasitas 100

func sendCallback(merchantURL string, transactionID string, statusCode int, message string) error {
	callbackData := CallbackJob{
		TransactionID: transactionID,
		StatusCode:    statusCode,
		Message:       message,
	}

	jsonData, err := json.Marshal(callbackData)
	if err != nil {
		return fmt.Errorf("failed to marshal callback data: %v", err)
	}

	resp, err := http.Post(merchantURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send callback: %v", err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response body: %v", err)
	}

	// Ambil hasil dari response
	callbackResult := fmt.Sprintf("%v", responseBody["result"]) // Sesuaikan dengan struktur response yang diharapkan

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

func sendCallbackWithRetry(merchantURL string, transactionID string, statusCode int, message string, retries int) {
	for i := 0; i < retries; i++ {
		err := sendCallback(merchantURL, transactionID, statusCode, message)
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
		sendCallbackWithRetry(job.MerchantURL, job.TransactionID, job.StatusCode, job.Message, 5)
	}
}
