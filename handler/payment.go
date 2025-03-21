package handler

import (
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
)

var TransactionCache = cache.New(5*time.Minute, 6*time.Minute)
var QrCache = cache.New(5*time.Minute, 10*time.Minute)

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

func PaymentQrisRedirect(c *fiber.Ctx) error {
	qrisUrl := c.Query("qrisUrl")
	acquirer := c.Query("acquirer")

	if qrisUrl == "" || acquirer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required parameters",
		})
	}

	// Buat ID transaksi unik (contoh pakai timestamp, bisa pakai UUID)
	transactionID := fmt.Sprintf("trx-%d", time.Now().UnixNano())

	// Simpan data di cache
	QrCache.Set(transactionID, qrisUrl+"|"+acquirer, cache.DefaultExpiration)

	// Redirect ke halaman tanpa query di URL
	return c.Redirect("/api/payment-qris/" + transactionID)
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

	if input.Amount > 500000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Amount exceeds the maximum allowed limit of 500000",
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

		var paymentMethod string
		switch inputReq.PaymentMethod {
		case "telkomsel_airtime_sms":
			paymentMethod = "telkomsel_airtime"
		case "telkomsel_airtime_ussd":
			paymentMethod = "telkomsel_airtime"
		case "xl_gcpay":
			paymentMethod = "xl_airtime"
		case "smartfren":
			paymentMethod = "smartfren_airtime"
		case "three":
			paymentMethod = "three_airtime"
		case "indosat_airtime_2":
			paymentMethod = "indosat_airtime"
		case "ovo_wallet":
			paymentMethod = "ovo"
		default:
			paymentMethod = inputReq.PaymentMethod

		}

		switch paymentMethod {
		case "xl_airtime":
			StrPaymentMethod = "XL"
		case "telkomsel_airtime":
			StrPaymentMethod = "Telkomsel"
		case "three_airtime":
			StrPaymentMethod = "Tri"
		case "smartfren_airtime":
			StrPaymentMethod = "Smartfren"
		case "indosat_airtime":
			StrPaymentMethod = "Indosat"
		case "shopeepay":
			StrPaymentMethod = "Shopeepay"
		case "gopay":
			StrPaymentMethod = "Gopay"
		case "qris":
			StrPaymentMethod = "Qris"
		case "va_bca":
			StrPaymentMethod = "BCA"
		case "dana":
			StrPaymentMethod = "Dana"
		case "ovo":
			StrPaymentMethod = "OVO"
		}

		if paymentMethod == "shopeepay" || paymentMethod == "gopay" || paymentMethod == "qris" || paymentMethod == "dana" || paymentMethod == "ovo" {
			vat := inputReq.Price - inputReq.Amount
			return c.Render("payment_ewallet", fiber.Map{
				"AppName":          inputReq.AppName,
				"PaymentMethod":    paymentMethod,
				"PaymentMethodStr": StrPaymentMethod,
				"ItemName":         inputReq.ItemName,
				"ItemId":           inputReq.ItemId,
				"Price":            inputReq.Price,
				"Amount":           inputReq.Amount,
				"Currency":         currency,
				"ClientAppKey":     inputReq.ClientAppKey,
				"VAT":              vat,
				"AppID":            inputReq.AppID,
				"MtID":             inputReq.MtTid,
				"UserId":           inputReq.UserId,
				"Token":            token,
				"BodySign":         inputReq.BodySign,
			})
		}

		return c.Render("payment", fiber.Map{
			"AppName":          inputReq.AppName,
			"PaymentMethod":    paymentMethod,
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

func QrisPage(c *fiber.Ctx) error {
	transactionID := c.Params("id")

	if transactionID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Transaction ID required")
	}

	// Ambil data dari cache
	data, found := QrCache.Get(transactionID)
	if !found {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found or expired")
	}

	// Pecah qrisUrl dan acquirer
	dataStr := data.(string)
	parts := strings.Split(dataStr, "|")
	if len(parts) != 2 {
		return c.Status(fiber.StatusInternalServerError).SendString("Invalid data format")
	}
	qrisUrl, acquirer := parts[0], parts[1]

	// Render halaman tanpa menampilkan query parameter
	return c.Render("payment_qris", fiber.Map{
		"QrisUrl":  qrisUrl,
		"Acquirer": acquirer,
	})
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
