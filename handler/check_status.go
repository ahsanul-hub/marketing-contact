package handler

import (
	"app/dto/http"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.elastic.co/apm"
)

type CheckStatusRequest struct {
	TransactionID string `json:"transaction_id"`
}

func CheckStatusDana(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := lib.CheckOrderDana(id)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func CheckStatusDanaFaspay(c *fiber.Ctx) error {
	id := c.Params("id")
	transaction, err := repository.GetTransactionByID(context.Background(), id)
	if err != nil || transaction == nil {
		return nil
	}

	res, err := lib.CheckOrderDanaFaspay(id, transaction.ReferenceID)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func CheckTransactionStatus(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CheckTransactionStatus", "handler")
	defer span.End()

	mtTid := c.Params("id")
	appKey := c.Get("appkey")
	appID := c.Get("appid")

	if mtTid == "" || appKey == "" || appID == "" {
		return response.Response(c, fiber.StatusBadRequest, "Missing required parameters")
	}

	transaction, err := repository.CheckTransactionByMerchantID(spanCtx, mtTid, appKey, appID)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, "Failed to get transaction: "+err.Error())
	}

	if transaction == nil {
		return response.Response(c, fiber.StatusNotFound, "Transaction not found")
	}

	var status string

	switch transaction.StatusCode {
	case 1000:
		status = "success"
	case 1003:
		status = "waiting send callback"
	case 1001:
		status = "pending"
	case 1005:
		status = "failed"
	}

	resp := http.TransactionStatus{
		UserID:                transaction.UserId,
		CreatedDate:           transaction.CreatedAt,
		MerchantTransactionID: transaction.MtTid,
		StatusCode:            transaction.StatusCode,
		PaymentMethod:         transaction.PaymentMethod,
		Amount:                fmt.Sprintf("%d", transaction.Amount),
		Status:                status,
		Currency:              transaction.Currency,
		ItemName:              transaction.ItemName,
		ItemID:                transaction.ItemId,
		ReferenceID:           transaction.ID,
	}

	return response.ResponseSuccess(c, fiber.StatusOK, resp)
}

func CheckStatusOvo(c *fiber.Ctx) error {
	id := c.Params("id")

	transaction, err := repository.GetTransactionByID(context.Background(), id)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	res, err := lib.CheckStatusOVO(transaction.ID, transaction.Amount, transaction.UserMDN, transaction.OvoBatchNo, transaction.OvoReferenceNumber)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func CheckTransactionStatusLegacy(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CheckTransactionStatusPost", "handler")
	defer span.End()

	var req CheckStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Response(c, fiber.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	appKey := c.Get("appkey")
	appID := c.Get("appid")

	if req.TransactionID == "" || appKey == "" || appID == "" {
		return response.Response(c, fiber.StatusBadRequest, "Missing required parameters")
	}

	transaction, err := repository.CheckTransactionByMerchantID(spanCtx, req.TransactionID, appKey, appID)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, "Failed to get transaction: "+err.Error())
	}

	if transaction == nil {
		return response.Response(c, fiber.StatusNotFound, "Transaction not found")
	}

	var status string
	var isSuccess bool
	var statusCode int
	switch transaction.StatusCode {
	case 1000:
		status = "payment_completed"
		statusCode = 1000
		isSuccess = true
	case 1003:
		status = "payment_completed"
		statusCode = 1000
		isSuccess = true
	case 1001:
		status = "pending"
		statusCode = 1001
		isSuccess = false
	case 1005:
		status = "failed"
		statusCode = 1005
		isSuccess = false
	default:
		isSuccess = false
		statusCode = transaction.StatusCode
		status = "unknown"
	}

	data := CheckStatusData{
		TransactionID: transaction.ID,
		UserMDN:       transaction.UserMDN,
		Amount:        transaction.Amount,
		ItemName:      transaction.ItemName,
		StatusCode:    fmt.Sprintf("%d", statusCode),
		Status:        status,
		Price:         fmt.Sprintf("%d", transaction.Price),
	}

	return c.JSON(fiber.Map{
		"success": isSuccess,
		"message": status,
		"data":    data,
	})
}

type CheckStatusData struct {
	TransactionID string `json:"_id"`
	UserMDN       string `json:"user_mdn"`
	Amount        uint   `json:"amount"`
	ItemName      string `json:"item_name"`
	StatusCode    string `json:"status_code"`
	Status        string `json:"status"`
	Price         string `json:"price"`
}
