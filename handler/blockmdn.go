package handler

import (
	"app/repository"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func BlockMDNHandler(c *fiber.Ctx) error {
	type Request struct {
		UserMDN  string `json:"user_mdn"`
		Duration *int64 `json:"duration,omitempty"` // Durasi dalam menit (opsional)
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input format",
		})
	}

	if req.UserMDN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "UserMDN is required",
		})
	}

	var duration *time.Duration
	if req.Duration != nil {
		d := time.Duration(*req.Duration) * time.Minute
		duration = &d
	}

	if err := repository.BlockMDN(req.UserMDN, duration); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to block MDN",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "MDN blocked successfully",
	})
}

// UnblockMDNHandler menangani request untuk membuka blokir MDN
func UnblockMDNHandler(c *fiber.Ctx) error {
	type Request struct {
		UserMDN string `json:"user_mdn"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input format",
		})
	}

	if req.UserMDN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "UserMDN is required",
		})
	}

	if err := repository.UnblockMDN(req.UserMDN); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to unblock MDN",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "MDN unblocked successfully",
	})
}

func BlockUserIdHandler(c *fiber.Ctx) error {
	type Request struct {
		UserID       string `json:"user_id"`
		MerchantName string `json:"merchant_name"`
		Duration     *int64 `json:"duration,omitempty"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input format",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "UserID is required",
		})
	}

	var duration *time.Duration
	if req.Duration != nil {
		d := time.Duration(*req.Duration) * time.Minute
		duration = &d
	}

	if err := repository.BlockUserID(req.UserID, req.MerchantName, duration); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to block userID",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "UserID blocked successfully",
	})
}

func UnblockUserIDHandler(c *fiber.Ctx) error {
	type Request struct {
		UserID       string `json:"user_id"`
		MerchantName string `json:"merchant_name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input format",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "UserID is required",
		})
	}

	if err := repository.UnblockUserID(req.UserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to unblock userID",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "userID unblocked successfully",
	})
}

func StartBlockedUserIDCacheRefresher() {
	go func() {
		// ticker untuk interval 15 menit
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		// pertama kali update saat start
		err := repository.UpdateBlockedUserIDCache()
		if err != nil {
			log.Printf("Gagal update BlockedUserIDCache saat startup: %v", err)
		}

		for {
			<-ticker.C
			err := repository.UpdateBlockedUserIDCache()
			if err != nil {
				log.Printf("Gagal update BlockedUserIDCache: %v", err)
			}
			log.Println("BlockedUserIDCache updated")
		}
	}()
}
