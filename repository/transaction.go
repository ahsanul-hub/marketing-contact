package repository

import (
	"app/database"
	"app/dto/model"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"math"
	"strconv"
	"strings"

	// "app/webhook"

	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/apm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// beautifyIDNumber menormalkan format nomor telepon
func beautifyIDNumber(mdn string, zero bool) string {
	check := true

	if mdn == "" {
		return ""
	}

	for check {
		check = false

		// Remove non-numeric prefix
		if len(mdn) > 0 && !isNumeric(string(mdn[0])) {
			mdn = mdn[1:]
			check = true
		}

		// Remove '62' prefix
		if strings.HasPrefix(mdn, "62") {
			mdn = mdn[2:]
			check = true
		}

		// Remove leading '0's
		for strings.HasPrefix(mdn, "0") {
			mdn = mdn[1:]
			check = true
		}
	}

	if zero {
		mdn = "0" + mdn
	} else {
		mdn = "62" + mdn
	}

	return mdn
}

// isNumeric checks if a string is numeric
func isNumeric(str string) bool {
	return str >= "0" && str <= "9"
}

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

func CreateTransaction(ctx context.Context, input *model.InputPaymentRequest, client *model.Client, appkey, appid string, vaBca *string) (string, uint, error) {
	span, _ := apm.StartSpan(ctx, "CreateTransaction", "repository")
	defer span.End()

	uniqueID, err := uuid.NewV7()
	if err != nil {
		log.Println("Error UUID:", err)
		return "", 0, err
	}

	fmt.Println("input:", input.UserIP)
	settlementConfig, err := GetSettlementConfig(client.UID)
	if err != nil {
		log.Println("Error GetSettlementConfig:", err)
	}

	var selectedSettlement *model.SettlementClient
	for _, settlement := range settlementConfig {
		if settlement.Name == input.PaymentMethod {
			selectedSettlement = &settlement
			break
		}
	}

	if selectedSettlement == nil {
		log.Println("selectedSettlement nil, check input.PaymentMethod:", input.PaymentMethod)
	}

	additionalPercent := 0.11
	if selectedSettlement != nil && selectedSettlement.AdditionalPercent != nil {
		additionalPercent = float64(*selectedSettlement.AdditionalPercent) / 100
	}

	chargingPrice := math.Ceil(float64(input.Amount)*additionalPercent + float64(input.Amount))
	nettSettlement := float64(input.Amount) * (float64(*selectedSettlement.SharePartner) / 100)

	currency := input.Currency
	if currency == "" {
		currency = "IDR"
	}

	transaction := model.Transactions{
		ID:                  uniqueID.String(),
		MtTid:               input.MtTid,
		StatusCode:          1001,
		ItemName:            input.ItemName,
		UserMDN:             input.UserMDN,
		Testing:             input.Testing,
		Route:               input.Route,
		UserId:              input.UserId,
		PaymentMethod:       input.PaymentMethod,
		Currency:            currency,
		Price:               uint(chargingPrice),
		NetSettlement:       float32(nettSettlement),
		Amount:              input.Amount,
		ItemId:              input.ItemId,
		BodySign:            input.BodySign,
		CustomerName:        input.CustomerName,
		UserIP:              input.UserIP,
		CallbackReferenceId: input.CallbackReferenceId,
	}

	transaction.AppID = appid
	transaction.MerchantName = client.ClientName
	transaction.AppName = client.AppName
	transaction.ClientAppKey = appkey
	transaction.NotificationUrl = input.NotificationUrl
	if vaBca != nil {
		transaction.VaBca = *vaBca
	}

	if err := database.DB.Create(&transaction).Error; err != nil {
		log.Println("Failed to create transaction:", err)
		return "", 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction.ID, transaction.Price, nil
}

func GetAllTransactions(
	ctx context.Context,
	limit, offset, status, denom int,
	transactionId, merchantTransactionId, userMDN, userId, appName string,
	appID, merchants, paymentMethods []string,
	startDate, endDate *time.Time,
) ([]model.Transactions, int64, error) {

	span, _ := apm.StartSpan(ctx, "GetAllTransactions", "repository")
	defer span.End()

	var transactions []model.Transactions
	var totalItems int64

	baseQuery := database.GetReadDB().Model(&model.Transactions{})

	if transactionId != "" {
		baseQuery = baseQuery.Where("id = ?", transactionId)
	}
	if merchantTransactionId != "" {
		baseQuery = baseQuery.Where("mt_tid = ?", merchantTransactionId)
	}
	if status != 0 {
		baseQuery = baseQuery.Where("status_code = ?", status)
	}
	if denom != 0 {
		baseQuery = baseQuery.Where("amount = ?", denom)
	}
	if userId != "" {
		baseQuery = baseQuery.Where("user_id = ?", userId)
	}
	if len(appID) > 0 {
		baseQuery = baseQuery.Where("app_id IN ?", appID)
	}
	if appName != "" {
		baseQuery = baseQuery.Where("app_name = ?", appName)
	}
	if len(merchants) > 0 {
		baseQuery = baseQuery.Where("merchant_name IN ?", merchants)
	}
	if userMDN != "" {
		baseQuery = baseQuery.Where("user_mdn = ?", userMDN)
	}
	if len(paymentMethods) > 0 {
		baseQuery = baseQuery.Where("payment_method IN ?", paymentMethods)
	}
	if startDate != nil && endDate != nil {
		baseQuery = baseQuery.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	countQuery := baseQuery.Session(&gorm.Session{})
	if err := countQuery.Count(&totalItems).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count transactions: %w", err)
	}

	dataQuery := baseQuery.Session(&gorm.Session{})
	if err := dataQuery.Debug().
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	return transactions, totalItems, nil
}

func GetTransactionsByDateRange(ctx context.Context, status int, startDate, endDate *time.Time, paymentMethod string, clientName, appID []string) ([]model.Transactions, error) {
	span, _ := apm.StartSpan(ctx, "GetTransactionsByDateRange", "repository")
	defer span.End()

	var transactions []model.Transactions
	query := database.GetReadDB()

	if status != 0 {
		query = query.Where("status_code = ?", status)
	}

	if len(clientName) > 0 {
		query = query.Where("merchant_name IN ?", clientName)
	}

	if len(appID) > 0 {
		query = query.Where("app_id IN ?", appID)
	}

	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}

	if startDate != nil && endDate != nil {
		startUTC := startDate.UTC()
		endUTC := endDate.UTC()

		query = query.Where("created_at BETWEEN ? AND ?", startUTC, endUTC)
	}

	if err := query.Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	return transactions, nil
}

func GetTransactionsMerchant(ctx context.Context, limit, offset, status, denom int, merchantTransactionId, clientName, userMDN, userId, appName string, paymentMethods []string, startDate, endDate *time.Time) ([]model.TransactionMerchantResponse, int64, error) {
	var transactions []model.Transactions
	query := database.GetReadDB()
	var totalItems int64

	if merchantTransactionId != "" {
		query = query.Where("mt_tid = ?", merchantTransactionId)
	}
	if status != 0 {
		query = query.Where("status_code = ?", status)
	}
	if denom != 0 {
		query = query.Where("amount = ?", denom)
	}
	if clientName != "" {
		query = query.Where("merchant_name = ?", clientName)
	}
	if appName != "" {
		query = query.Where("app_name = ?", appName)
	}
	if userMDN != "" {
		query = query.Where("user_mdn = ?", userMDN)
	}
	if userId != "" {
		query = query.Where("user_id = ?", userId)
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

	if err := query.Select("id, user_mdn, user_id, payment_method, mt_tid , status_code, amount, price, item_name, item_id,app_name, created_at, updated_at").Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to fetch transactions: %w", err)
	}

	var response []model.TransactionMerchantResponse
	for _, transaction := range transactions {
		response = append(response, model.TransactionMerchantResponse{
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

func GetAppNameFromClient(client *model.Client, clientID string) string {
	for _, app := range client.ClientApps {
		if app.AppID == clientID {
			return app.AppName
		}
	}
	return ""
}

func GetTransactionMerchantByID(ctx context.Context, appKey, appId, id string) (*model.TransactionMerchantResponse, error) {
	var transaction model.Transactions
	if err := database.DB.Where("id = ?", id).First(&transaction).Error; err != nil {
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
		FailReason:              transaction.FailReason,
		Price:                   transaction.Price,
		CreatedAt:               transaction.CreatedAt,
		UpdatedAt:               transaction.UpdatedAt,
	}

	return &response, nil
}

func UpdateTransactionStatusExpired(ctx context.Context, transactionID string, newStatusCode int, responseCallback, failReason string) error {
	db := database.DB

	callbackDate := time.Now()

	transactionUpdate := model.Transactions{
		StatusCode:          newStatusCode,
		FailReason:          failReason,
		ReceiveCallbackDate: &callbackDate,
	}
	timeLimit := time.Now().Add(-9 * time.Minute)

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? AND created_at <= ?", transactionID, timeLimit).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func UpdateOvoRefBatch(ctx context.Context, transactionID string, ovoBatch, ovoReference string) error {
	db := database.DB

	var transaction model.Transactions
	if err := db.WithContext(ctx).Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("transaction not found: %s", transactionID)
		}
		return fmt.Errorf("error fetching transaction: %w", err)
	}

	transaction.OvoBatchNo = ovoBatch
	transaction.OvoReferenceNumber = ovoReference

	if err := db.WithContext(ctx).Save(&transaction).Error; err != nil {
		return fmt.Errorf("failed to update ovo batch and ref number: %w", err)
	}

	return nil
}

func GetTransactionVa(ctx context.Context, vaNumber string) (*model.Transactions, error) {
	var transaction model.Transactions

	// Hitung batas waktu 70 menit yang lalu
	timeLimit := time.Now().Add(-70 * time.Minute)

	// Query berdasarkan va_bca dan CreatedAt dalam 70 menit terakhir
	if err := database.DB.WithContext(ctx).
		Where("va_bca = ? AND created_at >= ? AND status_code NOT IN ?", vaNumber, timeLimit, []int{1000, 1003}).
		First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}

	return &transaction, nil
}

func GetTransactionMoTelkomsel(ctx context.Context, msisdn, keyword string, otp int) (*model.Transactions, error) {
	var transaction model.Transactions

	// Samakan format MSISDN dengan yang disimpan di RequestMoTsel
	beautifyMsisdn := beautifyIDNumber(msisdn, true)

	// 1. Cek Redis dulu (jika tersedia)
	if database.RedisClient != nil {
		cacheKey := fmt.Sprintf("tsel:tx:%s:%s:%d", beautifyMsisdn, keyword, otp)
		// log.Println("cacheKey getMo ", cacheKey)
		val, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			// log.Printf("✅ Cache HIT untuk key: %s", cacheKey)
			// log.Printf("📦 Data dari Redis: %s", val)

			// Decode sebagai cacheData dari RequestMoTsel
			var cacheData struct {
				TransactionID string `json:"transaction_id"`
				Msisdn        string `json:"msisdn"`
				Keyword       string `json:"keyword"`
				Amount        string `json:"amount"`
				Otp           int    `json:"otp"`
				CreatedAt     int64  `json:"created_at"`
			}

			if err := json.Unmarshal([]byte(val), &cacheData); err == nil && cacheData.TransactionID != "" {
				// Parse amount dari cache (string -> uint)
				var amtUint uint
				if u, err := strconv.ParseUint(cacheData.Amount, 10, 64); err == nil {
					amtUint = uint(u)
				}

				transaction = model.Transactions{
					ID:         cacheData.TransactionID,
					UserMDN:    beautifyMsisdn,
					Keyword:    cacheData.Keyword,
					Otp:        cacheData.Otp,
					Amount:     amtUint,
					StatusCode: 1001, // Status pending
				}

				return &transaction, nil
			} else {
				log.Printf("❌ Gagal decode cacheData: %v", err)
			}
		}
	}

	// log.Printf("data tidak ada di redis untuk nomor: %s (beautify: %s) otp:%d ", msisdn, beautifyMsisdn, otp)

	// 2. Kalau tidak ada → query DB
	err := database.DB.WithContext(ctx).
		Where("user_mdn = ? AND keyword = ? AND otp = ? AND status_code = ?", msisdn, keyword, otp, 1001).
		First(&transaction).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 3. Simpan ke Redis (TTL misalnya 5 menit) - jika tersedia
	if database.RedisClient != nil {
		data, _ := json.Marshal(transaction)
		_ = database.RedisClient.Set(ctx, fmt.Sprintf("tsel:tx:%s:%s:%d", msisdn, keyword, otp), data, 5*time.Minute).Err()
	}

	return &transaction, nil
}

func UpdateTransactionStatus(ctx context.Context, transactionID string, newStatusCode int, referenceId, ximpayId *string, failReason string, receiveCallbackDate *time.Time) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		StatusCode: newStatusCode,
	}

	if ximpayId != nil {
		transactionUpdate.XimpayID = *ximpayId
	}
	if referenceId != nil {
		transactionUpdate.ReferenceID = *referenceId
	}

	if failReason != "" {
		transactionUpdate.FailReason = failReason
	}
	if receiveCallbackDate != nil {
		transactionUpdate.ReceiveCallbackDate = receiveCallbackDate
	}

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ? ", transactionID).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func UpdateTransactionKeyword(ctx context.Context, transactionID string, keyword string, otp int) error {
	db := database.DB

	transactionUpdate := model.Transactions{
		Keyword: keyword,
		Otp:     otp,
	}

	if err := db.WithContext(ctx).Model(&model.Transactions{}).Where("id = ?", transactionID).Updates(transactionUpdate).Error; err != nil {
		return fmt.Errorf("failed to update transaction keyword/otp: %w", err)
	}

	return nil
}

func GetPendingTransactions(ctx context.Context, paymentMethod string) ([]model.Transactions, error) {
	var transactions []model.Transactions

	// Optimasi query dengan menambahkan limit dan index yang tepat
	query := database.DB.WithContext(ctx).
		Select("id, merchant_name, status_code, created_at, reference_id").
		Where("status_code = ?", 1001).
		Order("created_at DESC"). // Ambil yang terbaru terlebih dahulu
		Limit(2000)               // Batasi hasil query untuk menghindari memory issues

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

func UpdateXimpayID(ctx context.Context, transactionID string, ximpayID string) error {
	db := database.DB

	var transaction model.Transactions
	if err := db.WithContext(ctx).Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("transaction not found: %s", transactionID)
		}
		return fmt.Errorf("error fetching transaction: %w", err)
	}

	transaction.XimpayID = ximpayID

	if err := db.WithContext(ctx).Save(&transaction).Error; err != nil {
		return fmt.Errorf("failed to update XimpayID: %w", err)
	}

	return nil
}

func UpdateMidtransId(ctx context.Context, transactionID string, midtransId string) error {
	db := database.DB

	var transaction model.Transactions
	if err := db.WithContext(ctx).Where("id = ?", transactionID).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("transaction not found: %s", transactionID)
		}
		return fmt.Errorf("error fetching transaction: %w", err)
	}

	transaction.MidtransTransactionId = midtransId

	if err := db.WithContext(ctx).Save(&transaction).Error; err != nil {
		return fmt.Errorf("failed to update Midtrans ID: %w", err)
	}

	return nil
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
