package handler

import (
	"app/dto/http"
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"context"
	"encoding/json"
	"log"
	"math"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
)

var TransactionCache = cache.New(4*time.Minute, 5*time.Minute)

func CreatePayment(c *fiber.Ctx) error {
	headers := map[string]string{
		"appkey":    c.Get("appkey"),
		"appid":     c.Get("appid"),
		"timestamp": c.Get("timestamp"),
		"nonce":     c.Get("nonce"),
		"secret":    c.Get("secret"),
		"bodysign":  c.Get("bodysign"),
	}

	arrClient, err := repository.FindClient(context.Background(), c.Get("appkey"), c.Get("appid"))

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	// log.Println(arrClient)

	var req http.CreatePaymentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validator.New().Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// checksecret() cek code legacy

	arrayTransactionCheck := CheckedTransaction(&req, arrClient)
	if !arrayTransactionCheck["success"].(bool) {
		return response.Response(c, fiber.StatusBadRequest, arrayTransactionCheck["retcode"].(string))
	}

	inputReq := model.InputPaymentRequest{
		ClientAppKey: headers["appkey"],
		Status:       helper.GetStatusMessage("1001"),
		Mobile:       arrayTransactionCheck["mobile"].(string),
		// Testing:       arrayTransactionCheck["testing"].(bool),
		Route:         arrayTransactionCheck["route"].(string),
		PaymentMethod: arrayTransactionCheck["payment_method"].(string),
		ItemName:      req.ItemName,
		Currency:      "IDR",
		Price:         arrayTransactionCheck["charging_price"].(uint),
	}

	// Beautify UserMDN
	if req.UserMDN != "" {
		req.UserMDN = helper.BeautifyIDNumber(strings.TrimSpace(inputReq.UserMDN), false)
	}

	// Create the transaction order
	transactionToken, err := repository.CreateOrder(context.Background(), &inputReq, arrClient)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, "E4001: Database Error")
	}

	// Save timestamps for transaction

	// Need check timestamp, check code legacy
	// err = SaveTransactionTimestamp(transactionToken)
	// if err != nil {
	// 	return response.Response(c, fiber.StatusInternalServerError, "E4001: Failed to update transaction timestamps")
	// }

	// Return successful response

	return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
		"token": transactionToken,
	})
}

func TestPayment(c *fiber.Ctx) error {
	// Mendapatkan data dari request body
	var requestData model.InputPaymentRequest
	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	err, _ := lib.SmartfrenTriyakomFlexible(requestData)

	// err, res := lib.SendData(requestData)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment successful",
		// "data:":   res,
	})
}

func CreateOrder(c *fiber.Ctx) error {
	var input model.InputPaymentRequest

	span, spanCtx := apm.StartSpan(c.Context(), "CreateOrderV1", "handler")
	defer span.End()

	// receivedBodysign := c.Get("bodysign")

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	if input.UserId == "" || input.MtTid == "" || input.PaymentMethod == "" || input.Amount == 0 || input.ItemName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Missing required fields in request body",
		})
	}

	arrClient, err := repository.FindClient(spanCtx, c.Get("appkey"), c.Get("appid"))

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	// bodyJSON, _ := json.Marshal(input)

	bodyJSON, err := json.Marshal(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error generating JSON",
		})
	}

	// Ubah bodyJSON menjadi string untuk dicetak
	bodyJSONString := string(bodyJSON)
	log.Println("bodyJSON:", bodyJSONString)

	appSecret := arrClient.ClientSecret
	log.Println("secret:", appSecret)

	expectedBodysign, _ := helper.GenerateBodySign(bodyJSONString, appSecret)
	log.Println("expectedBodysign", expectedBodysign)

	// if receivedBodysign != expectedBodysign {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Invalid bodysign",
	// 	})
	// }

	transactionID := uuid.New().String()

	amountFloat := float64(input.Amount)

	input.Price = uint(amountFloat + math.Round(0.11*amountFloat))
	input.AppID = c.Get("appid")
	input.ClientAppKey = c.Get("appkey")
	input.AppName = arrClient.ClientName

	TransactionCache.Set(transactionID, input, cache.DefaultExpiration)

	data := map[string]interface{}{
		"token": transactionID,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"retcode": "0000",
		"message": "Successful",
		"data":    data,
	})
}

func PaymentPage(c *fiber.Ctx) error {
	span, _ := apm.StartSpan(c.Context(), "PaymentPage", "handler")
	defer span.End()
	token := c.Params("token")

	if cachedData, found := TransactionCache.Get(token); found {
		inputReq := cachedData.(model.InputPaymentRequest)
		var StrPaymentMethod string

		currency := inputReq.Currency
		if currency == "" {
			currency = "IDR"
		}

		switch inputReq.PaymentMethod {
		case "xl_airtime":
			StrPaymentMethod = "XL"
		case "telkomsel_airtime":
			StrPaymentMethod = "Telkomsel"
		}

		// log.Println("inputreq:", inputReq)
		return c.Render("payment", fiber.Map{
			"AppName":          inputReq.AppName,
			"PaymentMethod":    inputReq.PaymentMethod,
			"PaymentMethodStr": StrPaymentMethod,
			"ItemName":         inputReq.ItemName,
			"Price":            inputReq.Price,
			"Amount":           inputReq.Amount,
			"Currency":         currency,
			"ClientAppKey":     inputReq.ClientAppKey,
			"AppID":            inputReq.AppID,
			"MtID":             inputReq.MtTid,
			"UserId":           inputReq.UserId,
			"RedirectURL":      inputReq.RedirectURL,
			"Token":            token,
		})

	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Transaction not found"})
}

func SuccessPage(c *fiber.Ctx) error {
	span, _ := apm.StartSpan(c.Context(), "SuccessPage", "handler")
	defer span.End()
	token := c.Params("token")
	msisdn := c.Params("msisdn")

	if cachedData, found := TransactionCache.Get(token); found {
		inputReq := cachedData.(model.InputPaymentRequest)
		var StrPaymentMethod string

		currency := inputReq.Currency
		if currency == "" {
			currency = "IDR"
		}

		switch inputReq.PaymentMethod {
		case "xl_airtime":
			StrPaymentMethod = "XL"
		case "telkomsel_airtime":
			StrPaymentMethod = "Telkomsel"
		}

		return c.Render("success_payment", fiber.Map{
			"PaymentMethodStr": StrPaymentMethod,
			"Msisdn":           msisdn,
			"RedirectURL":      inputReq.RedirectURL,
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Error"})
}
