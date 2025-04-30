package handler

import (
	"app/lib"
	"app/repository"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BillRequest struct {
	CompanyCode     string `json:"CompanyCode"`
	CustomerNumber  string `json:"CustomerNumber"`
	RequestID       string `json:"RequestID"`
	ChannelType     string `json:"ChannelType"`
	TransactionDate string `json:"TransactionDate"`
	AdditionalData  string `json:"AdditionalData,omitempty"`
}

type BillResponse struct {
	CompanyCode    string `json:"CompanyCode"`
	CustomerNumber string `json:"CustomerNumber"`
	RequestID      string `json:"RequestID"`
	InquiryStatus  string `json:"InquiryStatus"`
	InquiryReason  *struct {
		Indonesian string `json:"Indonesian,omitempty"`
		English    string `json:"English,omitempty"`
	} `json:"InquiryReason,omitempty"`
	CustomerName string `json:"CustomerName,omitempty"`
	CurrencyCode string `json:"CurrencyCode,omitempty"`
	TotalAmount  string `json:"TotalAmount,omitempty"`
	SubCompany   string `json:"SubCompany,omitempty"`
	DetailBills  *[]struct {
		BillDescription *struct {
			Indonesian string `json:"Indonesian,omitempty"`
			English    string `json:"English,omitempty"`
		} `json:"BillDescription,omitempty"`
		BillAmount     string `json:"BillAmount,omitempty"`
		BillNumber     string `json:"BillNumber,omitempty"`
		BillSubCompany string `json:"BillSubCompany,omitempty"`
	} `json:"DetailBills,omitempty"`
	FreeTexts *[]struct {
		Indonesian string `json:"Indonesian,omitempty"`
		English    string `json:"English,omitempty"`
	} `json:"FreeTexts,omitempty"`
	AdditionalData string `json:"AdditionalData,omitempty"`
}

type PaymentRequest struct {
	CompanyCode     string `json:"CompanyCode"`
	CustomerNumber  string `json:"CustomerNumber"`
	RequestID       string `json:"RequestID"`
	ChannelType     string `json:"ChannelType"`
	TransactionDate string `json:"TransactionDate"`
	AmountPaid      uint   `json:"PaidAmount"`
	AdditionalData  string `json:"AdditionalData,omitempty"`
}

// Struktur response untuk pembayaran
type PaymentResponse struct {
	CompanyCode       string `json:"CompanyCode"`
	CustomerNumber    string `json:"CustomerNumber"`
	RequestID         string `json:"RequestID"`
	PaymentFlagStatus string `json:"PaymentFlagStatus"`
	PaymentFlagReason *struct {
		Indonesian string `json:"Indonesian,omitempty"`
		English    string `json:"English,omitempty"`
	} `json:"PaymentFlagReason,omitempty"`
	CurrencyCode    string `json:"CurrencyCode,omitempty"`
	PaidAmount      string `json:"PaidAmount,omitempty"`
	TotalAmount     string `json:"TotalAmount,omitempty"`
	TransactionDate string `json:"TransactionDate,omitempty"`
}

type VaBCAErrorResponse struct {
	ErrorCode    string       `json:"ErrorCode"`
	ErrorMessage ErrorMessage `json:"ErrorMessage"`
}

type ErrorMessage struct {
	Indonesian string `json:"Indonesian"`
	English    string `json:"English"`
}

func InquiryBca(c *fiber.Ctx) error {

	authorization := c.Get("Authorization")
	x_bca_key := c.Get("X-BCA-Key")
	x_bca_signature := c.Get("X-BCA-Signature")
	x_bca_timestamp := c.Get("X-BCA-Timestamp")

	log.Println("authorization", authorization)
	log.Println("x_bca_key", x_bca_key)
	log.Println("x_bca_signature", x_bca_signature)
	log.Println("x_bca_timestamp", x_bca_timestamp)

	var request BillRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var response BillResponse

	vaNumber := fmt.Sprintf("11131%s", request.CustomerNumber)

	data, found := VaTransactionCache.Get(vaNumber)
	if !found {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found or expired")
	}

	// Pecah qrisUrl dan acquirer
	dataStr := data.(string)
	parts := strings.Split(dataStr, "|")
	if len(parts) != 3 {
		return c.Status(fiber.StatusInternalServerError).SendString("Invalid data format")
	}
	// transactionID := parts[0]

	transaction, err := repository.GetTransactionVa(context.Background(), vaNumber)
	if err != nil {
		response = BillResponse{
			CompanyCode:    request.CompanyCode,
			CustomerNumber: request.CustomerNumber,
			RequestID:      request.RequestID,
			InquiryStatus:  "01",
			InquiryReason: &struct {
				Indonesian string `json:"Indonesian,omitempty"`
				English    string `json:"English,omitempty"`
			}{
				Indonesian: "Nomor VA tidak valid atau expired",
				English:    "Invalid VA number or expired",
			},
			CurrencyCode: "IDR",
			// TotalAmount:  "150000.00",
			SubCompany: "00000",
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	totalAmount := fmt.Sprintf("%d.00", transaction.Amount)
	response = BillResponse{
		CompanyCode:    request.CompanyCode,
		CustomerNumber: request.CustomerNumber,
		RequestID:      request.RequestID,
		InquiryStatus:  "00",
		InquiryReason: &struct {
			Indonesian string `json:"Indonesian,omitempty"`
			English    string `json:"English,omitempty"`
		}{
			Indonesian: "Sukses",
			English:    "Success",
		},
		CurrencyCode: "IDR",
		TotalAmount:  totalAmount,
		SubCompany:   "00000",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func TokenBca(c *fiber.Ctx) error {

	authorization := c.Get("Authorization")
	expectedAuthorization := "Basic UjNkMXMxMG46YXRkc1Vxcml3MTQxQVQzTDlQNFo="

	var resError VaBCAErrorResponse

	if authorization != expectedAuthorization {
		resError = VaBCAErrorResponse{
			ErrorCode: "ERROR-INVALID-AUTHORIZATION",
			ErrorMessage: ErrorMessage{
				Indonesian: "client_id/client_secret tidak valid",
				English:    "Invalid client_id/client_secret",
			},
		}
		return c.Status(fiber.StatusOK).JSON(resError)
	}

	tokenRedpay, err := lib.RequestTokenVaBCARedpay()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error request token"})
	}

	return c.Status(fiber.StatusOK).JSON(tokenRedpay)
}

func PaymentBca(c *fiber.Ctx) error {
	var request PaymentRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	vaNumber := fmt.Sprintf("11131%s", request.CustomerNumber)

	transaction, err := repository.GetTransactionVa(context.Background(), vaNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Transaction not found or expired"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	timeLimit := time.Now().Add(-70 * time.Minute)
	if transaction.CreatedAt.Before(timeLimit) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Transaction expired"})
	}

	if request.AmountPaid != transaction.Amount {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payment amount"})
	}

	now := time.Now()

	receiveCallbackDate := &now

	err = repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, nil, nil, "", receiveCallbackDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update transaction status"})
	}

	// Format response
	response := PaymentResponse{
		CompanyCode:       request.CompanyCode,
		CustomerNumber:    request.CustomerNumber,
		RequestID:         request.RequestID,
		PaymentFlagStatus: "00",
		PaymentFlagReason: &struct {
			Indonesian string `json:"Indonesian,omitempty"`
			English    string `json:"English,omitempty"`
		}{
			Indonesian: "Pembayaran sukses",
			English:    "Payment successful",
		},
		CurrencyCode:    "IDR",
		PaidAmount:      fmt.Sprintf("%d.00", request.AmountPaid),
		TotalAmount:     fmt.Sprintf("%d.00", transaction.Amount),
		TransactionDate: time.Now().Format("2006-01-02 15:04:05"),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
