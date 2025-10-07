package repository

import (
	"app/database"
	"context"
	"errors"
	"log"

	// "app/lib"
	"app/dto/model"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

var BlockedMDNCache = cache.New(10*time.Minute, 15*time.Minute)
var BlockedUserIDCache = cache.New(10*time.Minute, 15*time.Minute)

type BlockedUserInfo struct {
	MerchantName string
	BlockedUntil *time.Time
}

func IsMDNBlocked(userMDN string) (bool, error) {
	// Query langsung ke database tanpa cache
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blocked model.BlockedMDN
	err := database.DB.WithContext(ctx).
		Where("user_mdn = ? AND (blocked_until IS NULL OR blocked_until > ?)", userMDN, time.Now()).
		First(&blocked).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// Consider blocked jika record ada dan belum expired (query sudah memfilter kondisi ini)
	return true, nil
}

func IsUserIDBlocked(userId, merchantName string) (bool, error) {
	if cached, found := BlockedUserIDCache.Get(userId); found {
		// Gunakan bool di cache: true = blocked, false = not blocked
		if v, ok := cached.(bool); ok {
			return v, nil
		}
		// Backward compatibility: jika masih ada tipe lama di cache, treat sebagai blocked hanya jika belum expired
		if v, ok := cached.(BlockedUserInfo); ok {
			if v.MerchantName == merchantName {
				if v.BlockedUntil == nil || time.Until(*v.BlockedUntil) > 0 {
					return true, nil
				}
			}
			// Jaga-jaga: convert ke negative cache 1 menit agar tidak hit tipe lama berulang
			BlockedUserIDCache.Set(userId, false, 1*time.Minute)
			return false, nil
		}
		log.Printf("unexpected cache type for userId %s: %T", userId, cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var blocked model.BlockedUserId
	err := database.DB.WithContext(ctx).
		Where("user_id = ? AND merchant_name = ? AND (blocked_until IS NULL OR blocked_until > ?)", userId, merchantName, time.Now()).
		First(&blocked).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// negative caching singkat: simpan false
			BlockedUserIDCache.Set(userId, false, 1*time.Minute)
			return false, nil
		}
		return false, err
	}

	// Set ke cache dengan TTL sesuai blocked_until
	var expiration time.Duration
	if blocked.BlockedUntil != nil {
		expiration = time.Until(*blocked.BlockedUntil)
		if expiration <= 0 {
			BlockedUserIDCache.Set(userId, BlockedUserInfo{MerchantName: merchantName, BlockedUntil: blocked.BlockedUntil}, 1*time.Minute)
			return false, nil
		}
	} else {
		expiration = cache.NoExpiration
	}

	// Simpan sebagai boolean true dengan TTL sesuai masa blokir
	BlockedUserIDCache.Set(userId, true, expiration)
	return true, nil
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
			expiration = time.Until(*mdn.BlockedUntil)
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
			expiration = time.Until(*user.BlockedUntil)
			if expiration <= 0 {
				continue
			}
		} else {
			expiration = cache.NoExpiration
		}

		// Simpan sebagai boolean true (blocked)
		BlockedUserIDCache.Set(user.UserId, true, expiration)
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
