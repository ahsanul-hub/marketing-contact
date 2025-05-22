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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
)

func CreateTransactionLegacy(c *fiber.Ctx) error {

	bodysign := c.Get("bodysign")
	appid := c.Params("appid")
	token := c.Params("token")

	if appid != "6078feb8764f1ba30a8b4569" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Endpoint",
		})
	}

	var transaction model.InputPaymentRequest

	cached, found := TransactionCache.Get(token)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction data not found in cache",
		})
	}

	cachedInput := cached.(model.InputPaymentRequest)

	transaction = model.InputPaymentRequest{
		RedirectURL:     cachedInput.RedirectURL,
		RedirectTarget:  cachedInput.RedirectTarget,
		UserId:          cachedInput.UserId,
		UserMDN:         cachedInput.UserMDN,
		MtTid:           cachedInput.MtTid,
		PaymentMethod:   cachedInput.PaymentMethod,
		Currency:        cachedInput.Currency,
		Amount:          cachedInput.Amount,
		ItemId:          cachedInput.ItemId,
		ItemName:        cachedInput.ItemName,
		ClientAppKey:    cachedInput.ClientAppKey,
		AppName:         cachedInput.AppName,
		AppID:           cachedInput.AppID,
		Status:          cachedInput.Status,
		BodySign:        cachedInput.BodySign,
		Mobile:          cachedInput.Mobile,
		Testing:         cachedInput.Testing,
		Route:           cachedInput.Route,
		Price:           cachedInput.Price,
		Otp:             cachedInput.Otp,
		ReffId:          cachedInput.ReffId,
		CustomerName:    cachedInput.CustomerName,
		NotificationUrl: cachedInput.NotificationUrl,
	}
	log.Println(cachedInput)

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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: customer name must not be empty",
		})
	}

	if !isEwallet && (transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN , item_name or amount must not be empty",
		})
	}

	beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Phone number %s is inactive or invalid, please try another number", transaction.UserMDN),
		})

	}

	if !isEwallet && !helper.IsValidPrefix(beautifyMsisdn, paymentMethod) && paymentMethod != "ovo" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid prefix, please use valid phone number.",
		})
	}

	arrClient, err := repository.FindClient(c.Context(), "xUkAmrJoE9C0XvUE8Di3570TT0FYwju4", appid)
	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	appName := transaction.AppName

	paymentMethodMap := make(map[string]model.PaymentMethodClient)
	for _, pm := range arrClient.PaymentMethods {
		paymentMethodMap[pm.Name] = pm
	}

	paymentMethodClient, exists := paymentMethodMap[paymentMethod]
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid payment method",
		})
	}

	var routes map[string][]string
	if err := json.Unmarshal(paymentMethodClient.Route, &routes); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Generate Va failed",
			})
		}

		vaNumber = res.VaNumber
		expiredTime = res.ExpiredTime
	}

	createdTransId, chargingPrice, err := repository.CreateTransaction(context.Background(), &transaction, arrClient, "xUkAmrJoE9C0XvUE8Di3570TT0FYwju4", appid, &vaNumber)
	if err != nil {
		log.Println("err", err)
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	switch paymentMethod {
	case "xl_airtime":

		validAmounts, exists := routes["xl_twt"]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No valid amounts found for the specified payment method",
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This denom is not supported for this payment method",
			})
		}

		_, err := lib.RequestChargingXL(transaction.UserMDN, transaction.MtTid, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		data := map[string]interface{}{
			"phone_number": transaction.UserMDN,
			"keyword":      "",
			"sms_code":     "",
			"short_code":   "99899",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})
	case "indosat_airtime":
		validAmounts, exists := routes["indosat_triyakom"]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No valid amounts found for the specified payment method",
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This denom is not supported for this payment method",
			})
		}

		ximpayId, err := lib.RequestChargingIsatTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request indosat failed",
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
			"short_code":   "",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "three_airtime":
		validAmounts, exists := routes["three_triyakom"]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No valid amounts found for the specified payment method",
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This denom is not supported for this payment method",
			})
		}

		ximpayId, err := lib.RequestChargingTriTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, transactionAmountStr)
		if err != nil {
			log.Println("Charging request tri failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "telkomsel_airtime":

		beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)
		_, keyword, otp, err := lib.RequestMoTsel(beautifyMsisdn, transaction.MtTid, transaction.ItemName, createdTransId, transactionAmountStr)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "99899",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "smartfren_airtime":
		validAmounts, exists := routes["smartfren_triyakom"]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No valid amounts found for the specified payment method",
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This denom is not supported for this payment method",
			})
		}

		ximpayId, err := lib.RequestChargingSfTriyakom(transaction.UserMDN, transaction.ItemName, createdTransId, chargingPrice)
		if err != nil {
			log.Println("Charging request smartfren failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "99899",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "shopeepay":

		res, err := lib.RequestChargingShopeePay(createdTransId, chargingPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request shopee failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "gopay":
		res, err := lib.RequestChargingGopay(createdTransId, chargingPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request gopay failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
			"token":                   token,
			"timestamp":               time.Now().Unix(),
			"merchant_transaction_id": transaction.MtTid,
		})

	case "qris":
		res, err := lib.RequestChargingQris(createdTransId, transaction.Amount)
		if err != nil {
			log.Println("Charging request qris failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
			"short_code":   "",
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
			"retcode":                 "0000",
			"message":                 "Successful",
			"data":                    data,
			"appid":                   appid,
			"appkey":                  arrClient.ClientAppkey,
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
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
