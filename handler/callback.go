package handler

import (
	"app/helper"
	"app/lib"
	"app/repository"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	msisdn := c.Query("msisdn")
	sms := c.Query("sms")
	trxId := c.Query("trx_id")

	// Pastikan sms memiliki format yang diharapkan
	parts := strings.Split(sms, " ")
	if len(parts) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid SMS format",
		})
	}

	keyword := strings.ToUpper(parts[0])
	otp, err := strconv.Atoi(parts[1])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid OTP format",
		})
	}

	// log.Println("Parsed keyword:", keyword, "OTP:", otp)

	beautifyMsisdn := helper.BeautifyIDNumber(msisdn, true)

	transaction, err := repository.GetTransactionMoTelkomsel(context.Background(), beautifyMsisdn, keyword, otp)
	if err != nil {
		log.Println("Error get transactions tsel:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	if transaction == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Transaction not found",
		})
	}

	denom := fmt.Sprintf("%d", transaction.Amount)
	res, err := lib.RequestMtTsel(transaction.UserMDN, trxId, denom)
	if err != nil {
		log.Println("Mt request failed:", err)
		return fmt.Errorf("Mt request failed:", err)
	}
	log.Println("res", res)

	if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, &trxId, nil, "", nil); err != nil {
		log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
	}

	return c.JSON(fiber.Map{
		"message": "MO request received",
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

func DanaCallback(c *fiber.Ctx) error {

	return nil
}

func CallbackHarsya(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing X-API-Key header",
		})
	}

	type Amount struct {
		Value    int    `json:"value"`
		Currency string `json:"currency"`
	}

	type HarsyaCallbackData struct {
		ID                string `json:"id"`
		ClientReferenceID string `json:"clientReferenceId"`
		Status            string `json:"status"`
		Amount            Amount `json:"amount"`
	}

	type HarsyaCallbackRequest struct {
		Event string             `json:"event"`
		Data  HarsyaCallbackData `json:"data"`
	}

	var req HarsyaCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	log.Println("reqCallbackHarsya", req)
	transactionID := req.Data.ClientReferenceID

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		log.Printf("Transaction not found: %s", transactionID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Transaction not found",
		})
	}

	now := time.Now()
	receiveCallbackDate := &now

	switch req.Data.Status {
	case "PROCESSING":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1001, nil, nil, "Processing payment", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to PROCESSING: %s", transactionID, err)
		}

	case "PAID":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "Payment completed", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to PAID: %s", transactionID, err)
		}

	case "CANCELLED":
		err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction cancelled")
		if err != nil {
			log.Printf("Error updating transaction %s to CANCELLED: %s", transactionID, err)
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Callback processed successfully",
	})
}

func ProcessUpdateTransactionPending() {

	for {
		transactions, err := repository.GetPendingTransactions(context.Background(), "telkomsel_airtime")

		if err != nil {
			log.Printf("Error retrieving pending transactions: %s", err)
		}

		for _, transaction := range transactions {
			if err != nil {
				log.Println("Error parsing CreatedAt for transaction:", transaction.ID, err)
				continue
			}

			createdAt := transaction.CreatedAt
			timeLimit := time.Now().Add(-15 * time.Minute)

			expired := createdAt.Before(timeLimit)

			if expired {
				if err := repository.UpdateTransactionStatusExpired(context.Background(), transaction.ID, 1005, "", "Transaction Expired"); err != nil {
					log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
				}
			}
		}

		time.Sleep(15 * time.Minute)
	}
}
