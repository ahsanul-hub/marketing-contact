package handler

import (
	"app/dto/http"
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"fmt"
	"html/template"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.elastic.co/apm"
)

var TransactionCache = cache.New(10*time.Minute, 11*time.Minute)
var VaTransactionCache = cache.New(60*time.Minute, 65*time.Minute)
var QrCache = cache.New(5*time.Minute, 10*time.Minute)

type CachedTransaction struct {
	Data      model.InputPaymentRequest
	IsClicked bool
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

func PaymentQrisRedirect(c *fiber.Ctx) error {
	qrisUrl := c.Query("qrisUrl")
	acquirer := c.Query("acquirer")
	backUrl := c.Query("back_url")

	if qrisUrl == "" || acquirer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required parameters",
		})
	}

	// Buat ID transaksi unik (contoh pakai timestamp, bisa pakai UUID)
	transactionID := fmt.Sprintf("trx-%d", time.Now().UnixNano())

	// Simpan data di cache
	QrCache.Set(transactionID, qrisUrl+"|"+acquirer+"|"+backUrl, cache.DefaultExpiration)

	// Redirect ke halaman tanpa query di URL
	return c.Redirect("/api/payment-qris/" + transactionID)
}

func CreateOrder(c *fiber.Ctx) error {
	var input model.InputPaymentRequest

	span, spanCtx := apm.StartSpan(c.Context(), "CreateOrderV1", "handler")
	defer span.End()
	appid := c.Get("appid")
	appkey := c.Get("appkey")

	receivedBodysign := c.Get("bodysign")

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	paymentLimits := map[string]uint{
		"qris":      10000000,
		"shopeepay": 10000000,
		"gopay":     10000000,
		"ovo":       10000000,
		"dana":      10000000,
	}

	limit, ok := paymentLimits[input.PaymentMethod]
	if !ok {
		limit = 500000
	}

	if input.Amount > limit {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Amount exceeds the maximum allowed limit of %d", limit),
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

	isBlocked, _ := repository.IsUserIDBlocked(input.UserId, arrClient.ClientName)
	if isBlocked {
		log.Println("userID is blocked")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "userID is blocked",
		})

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

	var paymentMethod string
	switch input.PaymentMethod {
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
	case "indosat_airtime2":
		paymentMethod = "indosat_airtime"
	case "ovo_wallet":
		paymentMethod = "ovo"
	case "smartfren_airtime2":
		paymentMethod = "smartfren_airtime"
	case "Three":
		paymentMethod = "three_airtime"
	case "Telkomsel":
		paymentMethod = "telkomsel_airtime"
	case "qr":
		paymentMethod = "qris"
	default:
		paymentMethod = input.PaymentMethod

	}

	settlementConfig, err := repository.GetSettlementConfig(arrClient.UID)
	if err != nil {
		log.Println("Error GetSettlementConfig:", err)
	}

	var selectedSettlement *model.SettlementClient
	for _, settlement := range settlementConfig {
		if settlement.Name == paymentMethod {
			selectedSettlement = &settlement
			break
		}
	}

	if selectedSettlement == nil {
		log.Println("selectedSettlement nil, check input.PaymentMethod:", paymentMethod)
	}

	input.Price = uint(amountFloat + math.Round(float64(*selectedSettlement.AdditionalPercent)/100*amountFloat))
	input.AppID = appid
	input.ClientAppKey = appkey
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
		case "indosat_airtime2":
			paymentMethod = "indosat_airtime"
		case "ovo_wallet":
			paymentMethod = "ovo"
		case "smartfren_airtime2":
			paymentMethod = "smartfren_airtime"
		case "Three":
			paymentMethod = "three_airtime"
		case "Telkomsel":
			paymentMethod = "telkomsel_airtime"
		case "qr":
			paymentMethod = "qris"
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
				"RedirectURL":      inputReq.RedirectURL,
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

	return c.Render("notfound", fiber.Map{})
}

func PayReturnSuccess(c *fiber.Ctx) error {
	return c.Render("payreturn_success", fiber.Map{})
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
	if len(parts) != 3 {
		return c.Status(fiber.StatusInternalServerError).SendString("Invalid data format")
	}

	qrisUrl, acquirer, backUrl := parts[0], parts[1], parts[2]

	// Render halaman tanpa menampilkan query parameter
	return c.Render("payment_qris", fiber.Map{
		"QrisUrl":     qrisUrl,
		"Acquirer":    acquirer,
		"RedirectURL": backUrl,
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
		case "ovo":
			StrPaymentMethod = "OVO"
			steps = []string{
				"Pastikan sudah login ke aplikasi OVO",
				`Pembayaran akan kadaluarsa dalam 55 detik.`,
				"Buka notifikasi OVO untuk melakukan pembayaran",
				`Pilih metode pembayaran "OVO Cash" atau "OVO Point" atau kombinasi keduanya, lalu klik "Bayar".`,
			}
		}

		if inputReq.PaymentMethod == "ovo" {
			return c.Render("success_ovo", fiber.Map{
				"PaymentMethodStr": StrPaymentMethod,
				"Msisdn":           msisdn,
				"RedirectURL":      inputReq.RedirectURL,
				"Steps":            steps,
			})
		}

		return c.Render("success_payment", fiber.Map{
			"PaymentMethodStr": StrPaymentMethod,
			"Msisdn":           msisdn,
			"RedirectURL":      inputReq.RedirectURL,
			"Steps":            steps,
		})
	}

	return c.Render("notfound", fiber.Map{})
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

func CreateTransactionVa(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CreateTransactionV1", "handler")
	defer span.End()

	bodysign := c.Get("bodysign")
	appkey := c.Get("appkey")
	appid := c.Get("appid")
	token := c.Get("token")

	var transaction model.InputPaymentRequest
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	if transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "" || transaction.CustomerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN , item_name or amount must not be empty",
		})
	}

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)

	appName := repository.GetAppNameFromClient(arrClient, appid)

	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	transaction.BodySign = bodysign
	arrClient.AppName = appName

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	res, err := lib.GenerateVA()
	if err != nil {
		log.Println("Generate va failed:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Generate Va failed",
		})
	}

	var transactionID string

	transactionID, _, err = repository.CreateTransaction(spanCtx, &transaction, arrClient, appkey, appid, &res.VaNumber)
	if err != nil {
		log.Println("err", err)
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	switch transaction.PaymentMethod {
	case "va_bca":
		vaPayment := http.VaPayment{
			VaNumber:      res.VaNumber,
			TransactionID: transactionID,
			CustomerName:  transaction.CustomerName,
			Bank:          "BCA",
			ExpiredDate:   res.ExpiredTime,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)
		TransactionCache.Delete(token)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"transaction_id": transactionID,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})

	case "va_bri":
		res, err := lib.VaHarsyaCharging(transactionID, transaction.CustomerName, "BRI", transaction.Amount)
		if err != nil {
			log.Println("Generate va failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Generate Va failed",
			})
		}

		// var vaPayment http.VaPayment

		// var bankName string

		// switch transaction.PaymentMethod {
		// case "va_bca":
		// 	bankName = "BCA"
		// case "va_bni":
		// 	bankName = "BNI"
		// case "va_bri":
		// 	bankName = "BRI"
		// case "va_mandiri":
		// 	bankName = "MANDIRI"
		// case "va_permata":
		// 	bankName = "PERMATA"
		// case "va_cimb":
		// 	bankName = "CIMB"
		// }

		vaPayment := http.VaPayment{
			VaNumber:      res.Data.ChargeDetails[0].VirtualAccount.VirtualAccountNumber,
			TransactionID: transactionID,
			Bank:          "BCA",
			ExpiredDate:   res.Data.ExpiryAt,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"transaction_id": transactionID,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})
	case "va_permata":
		res, err := lib.VaHarsyaCharging(transactionID, transaction.CustomerName, "PERMATA", transaction.Amount)
		if err != nil {
			log.Println("Generate va failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Generate Va failed",
			})
		}

		// var vaPayment http.VaPayment

		// var bankName string

		// switch transaction.PaymentMethod {
		// case "va_bca":
		// 	bankName = "BCA"
		// case "va_bni":
		// 	bankName = "BNI"
		// case "va_bri":
		// 	bankName = "BRI"
		// case "va_mandiri":
		// 	bankName = "MANDIRI"
		// case "va_permata":
		// 	bankName = "PERMATA"
		// case "va_cimb":
		// 	bankName = "CIMB"
		// }
		vaPayment := http.VaPayment{
			VaNumber:      res.Data.ChargeDetails[0].VirtualAccount.VirtualAccountNumber,
			TransactionID: transactionID,
			Bank:          "PERMATA",
			ExpiredDate:   res.Data.ExpiryAt,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"transaction_id": transactionID,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})
	case "va_mandiri":
		res, err := lib.VaHarsyaCharging(transactionID, transaction.CustomerName, "MANDIRI", transaction.Amount)
		if err != nil {
			log.Println("Generate va failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Generate Va failed",
			})
		}

		// var vaPayment http.VaPayment

		// var bankName string

		// switch transaction.PaymentMethod {
		// case "va_bca":
		// 	bankName = "BCA"
		// case "va_bni":
		// 	bankName = "BNI"
		// case "va_bri":
		// 	bankName = "BRI"
		// case "va_mandiri":
		// 	bankName = "MANDIRI"
		// case "va_permata":
		// 	bankName = "PERMATA"
		// case "va_cimb":
		// 	bankName = "CIMB"
		// }

		vaPayment := http.VaPayment{
			VaNumber:      res.Data.ChargeDetails[0].VirtualAccount.VirtualAccountNumber,
			TransactionID: transactionID,
			Bank:          "BCA",
			ExpiredDate:   res.Data.ExpiryAt,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"transaction_id": transactionID,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})
	case "va_bni":
		res, err := lib.VaHarsyaCharging(transactionID, transaction.CustomerName, "BNI", transaction.Amount)
		if err != nil {
			log.Println("Generate va failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Generate Va failed",
			})
		}

		// var vaPayment http.VaPayment

		// var bankName string

		// switch transaction.PaymentMethod {
		// case "va_bca":
		// 	bankName = "BCA"
		// case "va_bni":
		// 	bankName = "BNI"
		// case "va_bri":
		// 	bankName = "BRI"
		// case "va_mandiri":
		// 	bankName = "MANDIRI"
		// case "va_permata":
		// 	bankName = "PERMATA"
		// case "va_cimb":
		// 	bankName = "CIMB"
		// }

		vaPayment := http.VaPayment{
			VaNumber:      res.Data.ChargeDetails[0].VirtualAccount.VirtualAccountNumber,
			TransactionID: transactionID,
			CustomerName:  transaction.CustomerName,
			Bank:          "BCA",
			ExpiredDate:   res.Data.ExpiryAt,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"transaction_id": transactionID,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": "Failed generate Va Number",
	})
}

func VaPage(c *fiber.Ctx) error {
	vaNumber := c.Params("va")

	if vaNumber == "" {
		return c.Status(fiber.StatusBadRequest).SendString("vaNumber required")
	}

	data, found := VaTransactionCache.Get(vaNumber)
	if !found {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found or expired")
	}
	inputReq := data.(http.VaPayment)

	steps := map[string]map[string][]template.HTML{
		"BCA": {
			"ATM": []template.HTML{
				"Kunjungi ATM BCA terdekat.",
				"Pilih menu <b>Transaksi Lainnya > Transfer > ke Rek BCA Virtual Account</b>.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran.",
				"Klik Ya/Next/Oke untuk menyelesaikan transaksi.",
				"Simpan struk transaksi sebagai bukti pembayaran.",
			},
			"Mobile Banking": []template.HTML{
				"Lakukan LOG IN pada aplikasi BCA Mobile.",
				"Pilih m-BCA lalu masukkan KODE AKSES m-BCA.",
				"Pilih M-TRANSFER lalu BCA VIRTUAL ACCOUNT.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran dan masukkan PIN.",
				"Pembayaran SELESAI.",
			},
			"Internet Banking": []template.HTML{
				"Login ke KlikBCA.",
				"Pilih Pembayaran > Virtual Account.",
				"Pilih M-TRANSFER lalu BCA VIRTUAL ACCOUNT.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Periksa detail transaksi dan selesaikan pembayaran.",
			},
		},
		"BRI": {
			"ATM": []template.HTML{
				"Input kartu ATM dan PIN Anda",
				"Pilih menu <b>Transaksi Lainnya > Pembayaran > Pilih Menu Lain-lain > Pilih Menu BRIVA</b>.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran.",
				"Klik Ya/Next/Oke untuk menyelesaikan transaksi.",
				"Simpan struk transaksi sebagai bukti pembayaran.",
			},
			"Mobile Banking": []template.HTML{
				"Lakukan LOG IN pada aplikasi BRI Mobile.",
				"Pilih Mobile Banking BRI.",
				"Pilih Menu Pembayaran",
				"Pilih Menu BRIVA",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Klik Kirim",
				"Bukti bayar akan dikirim melalui sms atau notifikasi",
				"Pembayaran SELESAI.",
			},
			"Internet Banking": []template.HTML{
				"Login ke Internet Banking.",
				"Pilih Pembayaran > Pilih BRIVA.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Klik Kirim.",
				"Masukan Password",
				"Masukan mToken",
				"Klik Kirim",
				"Bukti bayar akan ditampilkan",
				"Selesai",
			},
		},
		"BNI": {
			"ATM": []template.HTML{
				"Input kartu ATM dan PIN Anda",
				"Pilih menu <b>Menu Lainnya > Transfer > Pilih Jenis rekening yang akan digunakan > Pilih “Virtual Account Billing”</b>.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran.",
				"Klik Ya/Next/Oke untuk menyelesaikan transaksi.",
				"Simpan struk transaksi sebagai bukti pembayaran.",
			},
			"Mobile Banking": []template.HTML{
				"Lakukan LOG IN pada aplikasi BNI Mobile Banking.",
				"Pilih menu Transfer.",
				"Pilih Menu Virtual Account Billing, kemudian pilih rekening debet.",
				// "Pilih Menu BRIVA",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Tagihan yang harus dibayarkan akan muncul pada layar konfirmasi",
				"Konfirmasi transaksi dan masukkan Password Transaksi",
				"Pembayaran SELESAI.",
			},
			"Internet Banking": []template.HTML{
				"Ketik alamat https://ibank.bni.co.id kemudian klik “Enter”",
				"Masukkan User ID dan Password.",
				"Pilih menu “Transfer”.",
				"Pilih “Virtual Account Billing”",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Kemudin tagihan yang harus dibayarkan akan muncul pada layar konfirmasi.",
				"Masukkan Kode Otentikasi Token.",
				"Pembayaran Anda telah berhasil",
			},
		},
		"MANDIRI": {
			"ATM": []template.HTML{
				"Pilih pembayaran/pembelian",
				"Pilih Multipayment",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran.",
				"Jika sudah benar masukkan angka/nomor 1, lalu pilih Ya",
				"Simpan struk transaksi sebagai bukti pembayaran.",
			},
			"Internet Banking": []template.HTML{
				"Masuk ke situs Bank Mandiri Internet Banking",
				"Login ke akun Mandiri Online",
				"Masuk ke menu “Pembayaran”",
				"Pilih menu “Multi Payment”",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Pilih “Lanjut”",
				"Konfirmasi pembayaran",
				"Masukkan PIN dan kode token",
			},
		},
		"PERMATA": {
			"ATM": []template.HTML{
				"Input kartu ATM dan PIN Anda",
				"Pilih menu <b> Pembayaran > Pembayaran > Pembayaran Lain > Virtual Account.",
				template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
				"Konfirmasi detail pembayaran.",
				"Klik Ya/Next/Oke untuk menyelesaikan transaksi.",
				"Simpan struk transaksi sebagai bukti pembayaran.",
			},
		},
	}

	// steps := map[string]map[string][]string{
	// 	"BCA": {
	// 		"ATM": {
	// 			"Kunjungi ATM BCA terdekat.",
	// 			"Pilih menu Transaksi Lainnya > Transfer > ke Rek BCA Virtual Account.",
	// 			template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
	// 			"Konfirmasi detail pembayaran.",
	// 			"Klik Ya/Next/Oke untuk menyelesaikan transaksi.",
	// 			"Simpan struk transaksi sebagai bukti pembayaran.",
	// 		},
	// 		"Mobile Banking": {
	// 			"Lakukan LOG IN pada aplikasi BCA Mobile.",
	// 			"Pilih m-BCA lalu masukkan KODE AKSES m-BCA.",
	// 			"Pilih M-TRANSFER lalu BCA VIRTUAL ACCOUNT.",
	// 			template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
	// 			"Konfirmasi detail pembayaran dan masukkan PIN.",
	// 			"Pembayaran SELESAI.",
	// 		},
	// 		"Internet Banking": {
	// 			"Login ke KlikBCA.",
	// 			"Pilih Transfer Dana > BCA Virtual Account.",
	// 			template.HTML("Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>"),
	// 			"Konfirmasi detail pembayaran.",
	// 			"Masukkan token KeyBCA dan selesaikan transaksi.",
	// 		},
	// 	},
	// 	"Mandiri": {
	// 		"ATM": {
	// 			"Masukkan kartu ATM dan PIN.",
	// 			"Pilih Bayar/Beli > Multi Payment.",
	// 			"Masukkan kode perusahaan dan nomor Virtual Account.",
	// 			"Periksa detail transaksi, lalu konfirmasi pembayaran.",
	// 		},
	// 		"Mobile Banking": {
	// 			"Login ke Livin' by Mandiri.",
	// 			"Pilih menu Bayar lalu Virtual Account.",
	// 			"Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>",
	// 			"Konfirmasi detail pembayaran dan selesaikan transaksi.",
	// 		},
	// 		"Internet Banking": {
	// 			"Login ke Mandiri Internet Banking.",
	// 			"Pilih Pembayaran > Virtual Account.",
	// 			"Masukkan nomor VA berikut: <b>" + inputReq.VaNumber + "</b>",
	// 			"Periksa detail transaksi dan selesaikan pembayaran.",
	// 		},
	// 	},
	// }

	return c.Render("va_page", fiber.Map{
		"VaNumber":     inputReq.VaNumber,
		"RedirectURL":  inputReq.RedirectURL,
		"CustomerName": inputReq.CustomerName,
		"BankName":     inputReq.Bank,
		"Steps":        steps[inputReq.Bank],
	})

}

func GetAllCachedTransactions(c *fiber.Ctx) error {
	transactions := make(map[string]CachedTransaction)

	TransactionCache.Items() // Mendapatkan semua item dalam cache
	for k, v := range TransactionCache.Items() {
		if cachedData, ok := v.Object.(CachedTransaction); ok {
			transactions[k] = cachedData
		}
	}

	return c.JSON(transactions)
}

// func MakePaid(c *fiber.Ctx) error {
// 	env := config.Config("ENV", "")
// 	id := c.Params("id")

// 	if env != "development" {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "This feature only for development",
// 		})
// 	}

// 	err := repository.UpdateTransactionStatus(context.Background(), id, 1003, nil, nil, "", nil)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "failed to update transaction",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message":       "transaction updated to paid",
// 		"transactionId": id,
// 	})
// }

// func MakeFailed(c *fiber.Ctx) error {
// 	env := config.Config("ENV", "")
// 	id := c.Params("id")

// 	if env != "development" {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "This feature only for development",
// 		})
// 	}

// 	err := repository.UpdateTransactionStatus(context.Background(), id, 1005, nil, nil, "", nil)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "failed to update transaction",
// 			"error":   err.Error(),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message":       "transaction updated to paid",
// 		"transactionId": id,
// 	})
// }
