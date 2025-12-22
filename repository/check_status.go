package repository

import (
	"app/database"
	"app/dto/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

func CheckTransactionByMerchantID(ctx context.Context, merchantTransactionID, appKey, appID string) (*model.Transactions, error) {

	var transaction model.Transactions

	// 1. Cek Redis
	if database.RedisClient != nil {
		cacheKey := fmt.Sprintf("tx:status:%s:%s", appID, merchantTransactionID)
		val, err := database.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(val), &transaction); err == nil {
				return &transaction, nil
			}
			log.Printf("⚠️ Failed to unmarshal cache for key %s: %v", cacheKey, err)
		}
	}

	// 2. Query ke DB (Gunakan ReadDB / Replica)
	if err := database.GetDB().WithContext(ctx).Where("mt_tid = ? AND client_app_key = ? AND app_id = ?", merchantTransactionID, appKey, appID).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}

	// 3. Simpan ke Redis (TTL 1 menit)
	if database.RedisClient != nil {
		cacheKey := fmt.Sprintf("tx:status:%s:%s", appID, merchantTransactionID)
		data, err := json.Marshal(transaction)
		if err == nil {
			_ = database.RedisClient.Set(ctx, cacheKey, data, 1*time.Minute).Err()
		} else {
			log.Printf("⚠️ Failed to marshal transaction for cache: %v", err)
		}
	}

	return &transaction, nil
}
