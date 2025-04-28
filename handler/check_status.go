package handler

import (
	"app/dto/http"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.elastic.co/apm"
)

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
