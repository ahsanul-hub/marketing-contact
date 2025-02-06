package handler

import (
	"app/repository"
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
)

func CallbackTriyakom(c *fiber.Ctx) error {
	// ximpayId := c.Query("ximpayid")
	ximpayStatus := c.Query("ximpaystatus")
	cbParam := c.Query("cbparam")
	// ximpaytoken := c.Query("ximpaytoken")
	// failcode := c.Query("failcode")
	transactionId := cbParam[1:]

	switch ximpayStatus {
	case "1":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionId, 1003, nil, ""); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionId, err)
		}
	case "2":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", "Insufficient balance"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transactionId, err)
		}
	case "3":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", "Undeliverable"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transactionId, err)
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Payment",
		})

	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Success",
	})
}
