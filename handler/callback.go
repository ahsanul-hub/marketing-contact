package handler

import (
	"app/repository"
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MidtransCallbackRequest struct {
	StatusCode        *string `json:"status_code"`
	TransactionID     *string `json:"transaction_id"`
	GrossAmount       *string `json:"gross_amount"`
	Currency          *string `json:"currency"`
	OrderID           *string `json:"order_id"`
	PaymentType       *string `json:"payment_type"`
	SignatureKey      *string `json:"signature_key"`
	TransactionStatus *string `json:"transaction_status"`
	FraudStatus       *string `json:"fraud_status"`
	StatusMessage     *string `json:"status_message"`
	MerchantID        *string `json:"merchant_id"`
	TransactionTime   *string `json:"transaction_time"`
	ExpiryTime        *string `json:"expiry_time"`
}

func CallbackTriyakom(c *fiber.Ctx) error {
	// ximpayId := c.Query("ximpayid")
	ximpayStatus := c.Query("ximpaystatus")
	cbParam := c.Query("cbparam")

	// ximpaytoken := c.Query("ximpaytoken")
	failcode := c.Query("failcode")
	transactionId := cbParam[1:]

	// log.Println("cbParam", cbParam)
	// log.Println("ximpayStatus", ximpayStatus)
	now := time.Now()

	receiveCallbackDate := &now

	switch ximpayStatus {
	case "1":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionId, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionId, err)
		}
	case "2":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", "Insufficient balance"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transactionId, err)
		}
	case "3":
		var failReason string

		switch failcode {
		case "301":
			failReason = "User not input phone number"
		case "302":
			failReason = "User not send sms"
		case "303":
			failReason = "User not input PIN"
		case "304":
			failReason = "Failed send PIN to user"
		case "305":
			failReason = "Over limit balance"
		case "306":
			failReason = "Failed Charge"
		case "307":
			failReason = "Failed send to operator"
		default:
			failReason = "Failed / Undeliverable"

		}
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", failReason); err != nil {
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

func MoTelkomsel(c *fiber.Ctx) error {
	rawInput := c.Queries()
	// msisdn := c.Query("msisdn")
	// sms := c.Query("sms")
	// adn := c.Query("adn")

	params := rawInput
	for k, v := range params {
		log.Println("key, v : ", k, v)
	}

	return c.JSON(fiber.Map{
		"message": "MO request received",
		"data":    rawInput,
	})
}

func MidtransCallback(c *fiber.Ctx) error {
	var req MidtransCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// log.Println("transactionId: ", *req.OrderID)
	// log.Println("transaction status: ", *req.TransactionStatus)

	if req.OrderID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields",
		})
	}

	var transactionID string

	transactionID = *req.OrderID

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		return nil
	}

	now := time.Now()

	receiveCallbackDate := &now

	switch *req.TransactionStatus {
	case "settlement":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", *req.TransactionID, err)
		}
	case "expire":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction expired"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", *req.TransactionID, err)
		}
	case "cancel", "deny", "failure":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction failed"); err != nil {
			log.Printf("Error updating transaction status for %s to failed: %s", *req.TransactionID, err)
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid transaction status",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Callback processed successfully",
	})
}
