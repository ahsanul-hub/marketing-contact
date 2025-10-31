package repository

import (
	"app/dto/model"
	"context"

	"gorm.io/gorm"
)

func InsertCreditCardLog(ctx context.Context, db *gorm.DB, log model.CreditCardLog) error {
	return db.WithContext(ctx).Create(&log).Error
}

func FindCreditCardLogsByFirst6(ctx context.Context, db *gorm.DB, first6 string) ([]model.CreditCardLog, error) {
	var logs []model.CreditCardLog
	err := db.WithContext(ctx).Where("first6 LIKE ?", first6+"%").Find(&logs).Error
	return logs, err
}
