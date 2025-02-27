package handler

import (
	"app/repository"
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
