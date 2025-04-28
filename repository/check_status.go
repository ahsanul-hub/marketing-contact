package repository

import (
	"app/database"
	"app/dto/model"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CheckTransactionByMerchantID(ctx context.Context, merchantTransactionID, appKey, appID string) (*model.Transactions, error) {

	var transaction model.Transactions
	if err := database.DB.Where("mt_tid = ? AND client_app_key = ? AND app_id = ?", merchantTransactionID, appKey, appID).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}
	return &transaction, nil
}
