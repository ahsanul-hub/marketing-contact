package handler

import (
	"app/dto/http"
	"app/dto/model"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"

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

func CheckStatusQrisHarsya(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := lib.CheckStatusHarsya(id)
	if err != nil {
		log.Println("err", err)
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

	if cached, found := TransactionCache.Get(req.TransactionID); found {
		cachedInput := cached.(model.InputPaymentRequestLegacy)

		var amount uint

		switch v := cachedInput.Amount.(type) {
		case float64:
			amount = uint(v)
		case float32:
			amount = uint(v)
		case int:
			amount = uint(v)
		case int64:
			amount = uint(v)
		case uint:
			amount = v
		case string:
			parsed, err := strconv.Atoi(v)
			if err == nil {
				amount = uint(parsed)
			} else {
				log.Println("Invalid amount string:", v)
				amount = 0
			}
		default:
			log.Println("Unknown amount type:", reflect.TypeOf(cachedInput.Amount))
			amount = 0
		}

		data := CheckStatusData{
			TransactionID: cachedInput.MtTid,
			UserMDN:       cachedInput.UserMDN,
			Amount:        amount,
			ItemName:      cachedInput.ItemName,
			StatusCode:    "1002",
			Status:        "waiting for payment",
			Price:         fmt.Sprintf("%d", cachedInput.Price),
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "waiting for payment",
			"data":    data,
		})
	}

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
