package repository

import (
	"app/database"
	"log"

	// "app/lib"
	"app/dto/model"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

var BlockedMDNCache = cache.New(10*time.Minute, 15*time.Minute)
var BlockedUserIDCache = cache.New(10*time.Minute, 15*time.Minute)

type BlockedUserInfo struct {
	MerchantName string
	BlockedUntil *time.Time
}

func IsMDNBlocked(userMDN string) (bool, error) {
	if cached, found := BlockedMDNCache.Get(userMDN); found {
		return cached.(bool), nil
	}
	return false, nil
}

func IsUserIDBlocked(userId, merchantName string) (bool, error) {
	if cached, found := BlockedUserIDCache.Get(userId); found {
		return cached.(bool), nil
	}

	if cached, found := BlockedUserIDCache.Get(userId); found {
		data := cached.(BlockedUserInfo)
		if data.MerchantName == merchantName {
			return true, nil
		}
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

func UpdateBlockedUserIDCache() error {
	var blockedUsers []model.BlockedUserId

	err := database.DB.Find(&blockedUsers).Error
	if err != nil {
		return err
	}

	BlockedUserIDCache.Flush()

	for _, user := range blockedUsers {
		var expiration time.Duration
		if user.BlockedUntil != nil {
			expiration = user.BlockedUntil.Sub(time.Now())
			if expiration <= 0 {
				continue
			}
		} else {
			expiration = cache.NoExpiration
		}

		BlockedUserIDCache.Set(user.UserId, BlockedUserInfo{
			MerchantName: user.MerchantName,
			BlockedUntil: user.BlockedUntil,
		}, expiration)
	}

	log.Println("Cache blocked UserID diperbarui dari database.")
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

func BlockUserID(userId, merchantName string, duration *time.Duration) error {

	var blockedUntil *time.Time
	if duration != nil {
		exp := time.Now().Add(*duration)
		blockedUntil = &exp
	}

	blockedUserID := model.BlockedUserId{
		UserId:       userId,
		MerchantName: merchantName,
		BlockedUntil: blockedUntil,
	}
	err := database.DB.Create(&blockedUserID).Error
	if err != nil {
		return err
	}

	var expiration time.Duration
	if duration != nil {
		expiration = *duration
	} else {
		expiration = cache.NoExpiration
	}
	BlockedUserIDCache.Set(userId, BlockedUserInfo{
		MerchantName: merchantName,
		BlockedUntil: blockedUntil,
	}, expiration)

	return nil
}

func UnblockUserID(userID string) error {

	err := database.DB.Where("user_id = ?", userID).Delete(&model.BlockedUserId{}).Error
	if err != nil {
		return err
	}

	BlockedUserIDCache.Delete(userID)

	return nil
}
