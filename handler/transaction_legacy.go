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
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

func CreateTransactionLegacy(c *fiber.Ctx) error {

	bodysign := c.Get("bodysign")
	appid := c.Params("appid")
	token := c.Params("token")
	clientIP := c.IP()

	allowedClients := map[string]string{
		"6078feb8764f1ba30a8b4569": "xUkAmrJoE9C0XvUE8Di3570TT0FYwju4",
		"64522e4e764f1bb11b8b4567": "1PSBWpSlKRY400bFIXKs2kBjNxLGf15h",
		"MHSBZnRBLkDQFlYDMSeXFA":   "5HjSLo37LwvIhTAX_zOJkg",
		"64d07790764f1bbe758b4569": "L66vZHbpCnCyjRzvnJ67wYeBEKPb5k1Q",
		"5ab32a23764f1b296b8bb386": "QdQpQLCBTbkAJv0OOTYhxAdojWkot5Gk",
	}

	expectedAppKey, exists := allowedClients[appid]
	if !exists {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0000",
			"message": "Unknown error",
			"data":    []interface{}{},
		})
	}
	var transaction model.InputPaymentRequest

	cached, found := TransactionCache.Get(token)
	if !found {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0000",
			"message": "Unknown error",
			"data":    []interface{}{},
		})
	}

	cachedInput := cached.(model.InputPaymentRequestLegacy)

	var amount uint

	switch v := cachedInput.Amount.(type) {
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0020",
				"message": "Invalid amount format!",
				"data":    []interface{}{},
			})
		}
		amount = uint(parsed)
	case float64:
		amount = uint(v)
	case int:
		amount = uint(v)
	case uint:
		amount = v
	default:
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0020",
			"message": "Unsupported amount type!",
			"data":    []interface{}{},
		})
	}

	transaction = model.InputPaymentRequest{
		RedirectURL:         cachedInput.RedirectURL,
		RedirectTarget:      cachedInput.RedirectTarget,
		UserId:              cachedInput.UserId,
		UserMDN:             cachedInput.UserMDN,
		MtTid:               cachedInput.MtTid,
		PaymentMethod:       cachedInput.PaymentMethod,
		Currency:            cachedInput.Currency,
		Amount:              amount,
		ItemId:              cachedInput.ItemId,
		ItemName:            cachedInput.ItemName,
		ClientAppKey:        cachedInput.ClientAppKey,
		AppName:             cachedInput.AppName,
		AppID:               cachedInput.AppID,
		Status:              cachedInput.Status,
		BodySign:            cachedInput.BodySign,
		Mobile:              cachedInput.Mobile,
		Testing:             cachedInput.Testing,
		Route:               cachedInput.Route,
		Price:               cachedInput.Price,
		Otp:                 cachedInput.Otp,
		ReffId:              cachedInput.ReffId,
		CustomerName:        cachedInput.CustomerName,
		NotificationUrl:     cachedInput.NotificationUrl,
		UserIP:              clientIP,
		CallbackReferenceId: token,
	}

	var paymentMethod string
	switch transaction.PaymentMethod {
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
		paymentMethod = transaction.PaymentMethod

	}

	var isEwallet bool

	if paymentMethod == "shopeepay" || paymentMethod == "gopay" || paymentMethod == "qris" || paymentMethod == "dana" || paymentMethod == "va_bca" {
		isEwallet = true
	}

	if paymentMethod == "va_bca" && (transaction.CustomerName == "") {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0013",
			"message": "Some field(s) missing",
			"data":    []interface{}{},
		})
	}

	if !isEwallet && (transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "") {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0013",
			"message": "Some field(s) missing",
			"data":    []interface{}{},
		})
	}

	// Validasi limit harian telco
	if isTelcoMethod(paymentMethod) {
		msisdnKey := helper.BeautifyIDNumber(transaction.UserMDN, true)
		ok, err := checkDailyTelcoLimit(msisdnKey, transaction.Amount)
		if err != nil {
			log.Println("error check telco limit:", err)
		}
		if !ok {
			log.Println("This number has exceeded the daily transaction limit with merchant_transaction_id:", transaction.MtTid)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "This number has exceeded the daily transaction limit",
			})
		}
	}

	beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0016",
			"message": "Invalid MSISDN!",
			"data":    []interface{}{},
		})

	}

	isBlockedMDN, err := repository.IsMDNBlocked(beautifyMsisdn)
	if err != nil {
		log.Println("error get blocked Msisdn:", err)

	}

	if isBlockedMDN {
		log.Println("diblokir: ", beautifyMsisdn)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Msisdn is blocked",
		})
	}

	if !isEwallet && !helper.IsValidPrefix(beautifyMsisdn, paymentMethod) && paymentMethod != "ovo" {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0016",
			"message": "Invalid MSISDN!",
			"data":    []interface{}{},
		})
	}

	arrClient, err := repository.FindClient(c.Context(), expectedAppKey, appid)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0001",
			"message": "Invalid appkey or appid",
			"data":    []interface{}{},
		})
	}

	appName := transaction.AppName

	paymentMethodMap := make(map[string]model.PaymentMethodClient)
	for _, pm := range arrClient.PaymentMethods {
		paymentMethodMap[pm.Name] = pm
	}

	paymentMethodClient, exists := paymentMethodMap[paymentMethod]
	if !exists {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0007",
			"message": "This payment method is not available for this merchant",
			"data":    []interface{}{},
		})
	}

	var routes map[string][]string
	if err := json.Unmarshal(paymentMethodClient.Route, &routes); err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0007",
			"message": "This payment method is not available for this merchant",
			"data":    []interface{}{},
		})
	}

	transactionAmountStr := fmt.Sprintf("%d", transaction.Amount)
	transaction.BodySign = bodysign
	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	arrClient.AppName = appName
	transaction.PaymentMethod = paymentMethod

	var vaNumber, expiredTime string

	if paymentMethod == "va_bca" {
		res, err := lib.GenerateVA()
		if err != nil {
			log.Println("Generate va failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Generate va failed",
				"data":    []interface{}{},
			})
		}

		vaNumber = res.VaNumber
		expiredTime = res.ExpiredTime
	}

	createdTransId, chargingPrice, err := repository.CreateTransaction(context.Background(), &transaction, arrClient, expectedAppKey, appid, &vaNumber)
	if err != nil {
		log.Println("err", err)
		return c.JSON(fiber.Map{
			"success": false,
			"retcode": "E0000",
			"message": "Failed create Transaction",
			"data":    []interface{}{},
		})
	}

	switch paymentMethod {
	case "xl_airtime", "xl_twt":

		validAmounts, exists := routes["xl_twt"]
		if !exists {

			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		validAmount := false
		for _, route := range validAmounts {
			if transactionAmountStr == route {
				validAmount = true
				break
			}
		}

		if !validAmount && !paymentMethodClient.Flexible {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		_, err := lib.RequestChargingXL(transaction.UserMDN, transaction.MtTid, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "99899",
			"url":          "",
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "reply",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"error_message":           "",
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})
	case "indosat_airtime", "indosat_triyakom":
		validAmounts, exists := routes["indosat_triyakom"]
		if !exists {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		validAmount := false
		for _, route := range validAmounts {
			if transactionAmountStr == route {
				validAmount = true
				break
			}
		}

		if !validAmount && !paymentMethodClient.Flexible {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		ximpayId, err := lib.RequestChargingIsatTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request failed:", err)

			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateXimpayID(context.Background(), createdTransId, ximpayId)
		if err != nil {
			log.Println("Updated Ximpay ID error:", err)
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          "",
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "reply",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "three_airtime", "three_triyakom":
		validAmounts, exists := routes["three_triyakom"]
		if !exists {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		validAmount := false
		for _, route := range validAmounts {
			if transactionAmountStr == route {
				validAmount = true
				break
			}
		}

		if !validAmount && !paymentMethodClient.Flexible {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		ximpayId, err := lib.RequestChargingTriTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, transactionAmountStr)
		if err != nil {
			log.Println("Charging request tri failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateXimpayID(context.Background(), createdTransId, ximpayId)
		if err != nil {
			log.Println("Updated Ximpay ID error:", err)
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          "",
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "reply",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "telkomsel_airtime", "telkomsel_airtime_sms":

		beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)
		_, keyword, otp, err := lib.RequestMoTsel(beautifyMsisdn, transaction.MtTid, transaction.ItemName, createdTransId, transactionAmountStr)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateTransactionKeyword(context.Background(), createdTransId, keyword, otp)
		if err != nil {
			log.Println("Updated Transaction Keyword error:", err)
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      keyword,
			"sms_code":     fmt.Sprintf("%d", otp),
			"shortcode":    "99899",
			"url":          "",
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "send_otp",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "smartfren_airtime", "smartfren_triyakom":
		validAmounts, exists := routes["smartfren_triyakom"]
		if !exists {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		validAmount := false
		for _, route := range validAmounts {
			if transactionAmountStr == route {
				validAmount = true
				break
			}
		}

		if !validAmount && !paymentMethodClient.Flexible {
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0008",
				"message": "This denom is not supported for this payment method",
				"data":    []interface{}{},
			})
		}

		ximpayId, err := lib.RequestChargingSfTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request smartfren failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateXimpayID(context.Background(), createdTransId, ximpayId)
		if err != nil {
			log.Println("Updated Ximpay ID error:", err)
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          "",
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "reply",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "shopeepay", "shopeepay_midtrans":

		res, err := lib.RequestChargingShopeePay(createdTransId, chargingPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request shopee failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateMidtransId(context.Background(), createdTransId, res.TransactionID)
		if err != nil {
			log.Println("Updated Midtrans ID error:", err)
		}

		data := map[string]interface{}{
			"ack":          "",
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          res.Actions[0].URL,
			"qr":           "",
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "redirect",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "gopay", "gopay_midtrans":
		res, err := lib.RequestChargingGopay(createdTransId, chargingPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request gopay failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateMidtransId(context.Background(), createdTransId, res.TransactionID)
		if err != nil {
			log.Println("Updated Midtrans ID error:", err)
		}

		data := map[string]interface{}{
			"ack":          "",
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          res.Actions[1].URL,
			"qr":           res.Actions[0].URL,
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "redirect",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "qris", "qris_midtrans":
		res, err := lib.RequestChargingQris(createdTransId, transaction.Amount)
		if err != nil {
			log.Println("Charging request qris failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		err = repository.UpdateMidtransId(context.Background(), createdTransId, res.TransactionID)
		if err != nil {
			log.Println("Updated Midtrans ID error:", err)
		}

		data := map[string]interface{}{
			"ack":          "",
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"shortcode":    "",
			"url":          res.Actions[0].URL,
			"qr":           res.QrString,
			"html":         "",
			"bankcode":     "",
			"vaid":         "",
			"code":         "",
			"trx_type":     "redirect",
		}

		return c.JSON(fiber.Map{
			"success":                 true,
			"data":                    data,
			"appid":                   appid,
			"appkey":                  expectedAppKey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})
	case "ovo":
		resultChan := make(chan *lib.OVOResponse)
		errorChan := make(chan error)

		// Jalankan ChargingOVO secara async
		go func() {
			res, err := lib.ChargingOVO(createdTransId, chargingPrice, transaction.UserMDN)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- res
		}()

		go func() {
			select {
			case err := <-errorChan:
				log.Println("Charging request ovo failed:", err)
				// Kamu bisa retry atau simpan log error ke DB jika perlu
				_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, nil, nil, "Charging request failed", nil)
			case res := <-resultChan:
				log.Println("res charge ovo:", res)

				referenceId := fmt.Sprintf("%s-%s", res.ApprovalCode, res.TransactionRequestData.MerchantInvoice)
				now := time.Now()
				receiveCallbackDate := &now

				switch res.ResponseCode {
				case "00":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1003, &referenceId, nil, "", receiveCallbackDate)
				case "13":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Invalid amount", receiveCallbackDate)
				case "14":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Invalid Mobile Number / OVO ID", receiveCallbackDate)
				case "17":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Decline", receiveCallbackDate)
				case "25":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Not Found", receiveCallbackDate)
				case "26":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Failed", receiveCallbackDate)
				case "40":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Failed", receiveCallbackDate)
				case "68":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Pending / Timeout", receiveCallbackDate)
				case "94":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Duplicate Request Params", receiveCallbackDate)
				case "ER":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "System Failure", receiveCallbackDate)
				case "EB":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Terminal Blocked", receiveCallbackDate)
				case "BR":
					_ = repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Bad Request", receiveCallbackDate)
				}
			}
		}()

		return c.JSON(fiber.Map{
			"success":  true,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
			"guide": fiber.Map{
				"en": "Please open the OVO application to continue payment.",
				"id": "Silahkan buka aplikasi OVO untuk melanjutkan pembayaran.",
			},
		})
	case "dana":
		strPrice := fmt.Sprintf("%d00", chargingPrice)
		res, err := lib.RequestChargingDanaFaspay(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL, transaction.CustomerName, transaction.UserMDN) //lib.RequestChargingDana(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request dana failed:", err)
			return c.JSON(fiber.Map{
				"success": false,
				"retcode": "E0000",
				"message": "Failed charging request",
				"data":    []interface{}{},
			})
		}

		if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1001, &res.TrxID, nil, "", nil); err != nil {
			log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
		}

		// log.Println("redirect: ", res.Actions[0].URL)
		return c.JSON(fiber.Map{
			"success":  true,
			"back_url": transaction.RedirectURL,
			"redirect": res.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
		})
	case "va_bca":
		vaPayment := http.VaPayment{
			VaNumber:      vaNumber,
			CustomerName:  transaction.CustomerName,
			TransactionID: createdTransId,
			Bank:          "BCA",
			ExpiredDate:   expiredTime,
		}

		VaTransactionCache.Set(vaPayment.VaNumber, vaPayment, cache.DefaultExpiration)

		return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
			"success":        true,
			"va":             vaPayment.VaNumber,
			"expired_date":   expiredTime,
			"customer_name":  transaction.CustomerName,
			"transaction_id": createdTransId,
			"retcode":        "0000",
			"message":        "Successful Created Transaction",
		})
	}

	return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"retcode": "0000",
		"message": "Successful Created Transaction",
	})
}
