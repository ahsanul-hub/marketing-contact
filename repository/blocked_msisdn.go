package repository

import (
	"app/database"

	// "app/lib"
	"app/dto/model"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

var BlockedMDNCache = cache.New(10*time.Minute, 15*time.Minute)

func IsMDNBlocked(userMDN string) (bool, error) {
	if cached, found := BlockedMDNCache.Get(userMDN); found {
		return cached.(bool), nil
	}
	return false, nil
}

func UpdateBlockedMDNCache() error {
	var blockedMDNs []model.BlockedMDN

	err := database.DB.Find(&blockedMDNs).Error
	if err != nil {
		return err
	}

	BlockedMDNCache.Flush()

	for _, mdn := range blockedMDNs {
		var expiration time.Duration
		if mdn.BlockedUntil != nil {
			expiration = mdn.BlockedUntil.Sub(time.Now())
			if expiration <= 0 {
				continue
			}
		} else {
			expiration = cache.NoExpiration
		}

		BlockedMDNCache.Set(mdn.UserMDN, true, expiration)
	}

	fmt.Println("Cache blocked MDN diperbarui dari database.")
	return nil
}

func BlockMDN(userMDN string, duration *time.Duration) error {

	var blockedUntil *time.Time
	if duration != nil {
		exp := time.Now().Add(*duration)
		blockedUntil = &exp
	}

	blockedMDN := model.BlockedMDN{
		UserMDN:      userMDN,
		BlockedUntil: blockedUntil,
	}
	err := database.DB.Create(&blockedMDN).Error
	if err != nil {
		return err
	}

	var expiration time.Duration
	if duration != nil {
		expiration = *duration
	} else {
		expiration = cache.NoExpiration
	}
	BlockedMDNCache.Set(userMDN, true, expiration)

	return nil
}

func UnblockMDN(userMDN string) error {

	err := database.DB.Where("user_mdn = ?", userMDN).Delete(&model.BlockedMDN{}).Error
	if err != nil {
		return err
	}

	BlockedMDNCache.Delete(userMDN)

	return nil
}
