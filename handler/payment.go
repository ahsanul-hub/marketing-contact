package handler

import (
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
)

var TransactionCache = cache.New(5*time.Minute, 6*time.Minute)

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

	receivedBodysign := c.Get("bodysign")

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

	// bodyJSON, err := json.Marshal(input)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Error generating JSON",
	// 	})
	// }

	// Ubah bodyJSON menjadi string untuk dicetak
	// bodyJSONString := string(bodyJSON)
	// log.Println("bodyJSON:", bodyJSONString)

	appSecret := arrClient.ClientSecret

	expectedBodysign := helper.GenerateBodySign(input, appSecret)
	// log.Println("expectedBodysign", expectedBodysign)

	if receivedBodysign != expectedBodysign {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid bodysign",
		})
	}

	transactionID := uuid.New().String()

	amountFloat := float64(input.Amount)

	input.Price = uint(amountFloat + math.Round(0.11*amountFloat))
	input.AppID = c.Get("appid")
	input.ClientAppKey = c.Get("appkey")
	input.AppName = arrClient.ClientName
	input.BodySign = receivedBodysign

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
		case "tri_airtime":
			StrPaymentMethod = "Tri"
		case "smartfren_airtime":
			StrPaymentMethod = "Smartfren"
		case "indosat_airtime":
			StrPaymentMethod = "Indosat"
		}

		// log.Println("inputreq:", inputReq)
		return c.Render("payment", fiber.Map{
			"AppName":          inputReq.AppName,
			"PaymentMethod":    inputReq.PaymentMethod,
			"PaymentMethodStr": StrPaymentMethod,
			"ItemName":         inputReq.ItemName,
			"ItemId":           inputReq.ItemId,
			"Price":            inputReq.Price,
			"Amount":           inputReq.Amount,
			"Currency":         currency,
			"ClientAppKey":     inputReq.ClientAppKey,
			"AppID":            inputReq.AppID,
			"MtID":             inputReq.MtTid,
			"UserId":           inputReq.UserId,
			"RedirectURL":      inputReq.RedirectURL,
			"Token":            token,
			"BodySign":         inputReq.BodySign,
		})

	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Transaction not found"})
}

func InputOTPSF(c *fiber.Ctx) error {
	span, _ := apm.StartSpan(c.Context(), "OtpPage", "handler")
	defer span.End()
	// ximpaytoken := c.Get("ximpaytoken")
	ximpayid := c.Params("ximpayid")
	token := c.Params("token")

	return c.Render("paymentotp", fiber.Map{
		"ReffId":           ximpayid,
		"Token":            token,
		"PaymentMethodStr": "Smartfren",
	})

}

func SuccessPage(c *fiber.Ctx) error {
	span, _ := apm.StartSpan(c.Context(), "SuccessPage", "handler")
	defer span.End()
	token := c.Params("token")
	msisdn := c.Params("msisdn")

	if cachedData, found := TransactionCache.Get(token); found {
		inputReq := cachedData.(model.InputPaymentRequest)
		var StrPaymentMethod string
		var steps []string

		currency := inputReq.Currency
		if currency == "" {
			currency = "IDR"
		}

		switch inputReq.PaymentMethod {
		case "xl_airtime":
			StrPaymentMethod = "XL"
			steps = []string{
				"Cek SMS yang masuk ke nomor anda",
				"Cek kembali informasi yang diterima di sms, kemudian balas sms dengan kode OTP yang diterima",
				"Pastikan pulsa cukup sesuai nominal transaksi",
				"Transaksi akan diproses setelah OTP dikirim.",
			}
		case "telkomsel_airtime":
			StrPaymentMethod = "Telkomsel"
		}

		return c.Render("success_payment", fiber.Map{
			"PaymentMethodStr": StrPaymentMethod,
			"Msisdn":           msisdn,
			"RedirectURL":      inputReq.RedirectURL,
			"Steps":            steps,
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Error"})
}

func SuccessPageOTP(c *fiber.Ctx) error {
	span, _ := apm.StartSpan(c.Context(), "SuccessPage", "handler")
	defer span.End()
	token := c.Params("token")

	if cachedData, found := TransactionCache.Get(token); found {
		inputReq := cachedData.(model.InputPaymentRequest)
		var StrPaymentMethod string

		switch inputReq.PaymentMethod {
		case "xl_airtime":
			StrPaymentMethod = "XL"
		case "telkomsel_airtime":
			StrPaymentMethod = "Telkomsel"
		case "smartfren_airtime":
			StrPaymentMethod = "Smartfren"
		}
		return c.Render("success_payment_otp", fiber.Map{
			"PaymentMethodStr": StrPaymentMethod,
			"RedirectURL":      inputReq.RedirectURL,
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": "Error"})
}
