package handler

import (
	"app/config"
	"app/dto/http"
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/pkg/response"
	"app/repository"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"github.com/xuri/excelize/v2"
	"go.elastic.co/apm"
)

func contains(denom []float64, amount float64) bool {
	for _, d := range denom {
		if d == amount {
			return true
		}
	}
	return false
}

type CallbackMerchantJob struct {
	CallbackUrl   string
	TransactionID string
	StatusCode    int
	Message       string
}

func containsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func CreateTransaction(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CreateTransactionV2", "handler")
	defer span.End()

	bodysign := c.Get("bodysign")
	appkey := c.Get("appkey")
	appid := c.Get("appid")
	receivedBodysign := c.Get("bodysign")

	var transaction model.InputPaymentRequest
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	mtDupKey := fmt.Sprintf("dup:%s:%s", appkey, transaction.MtTid)

	if _, found := MTIDCache.Get(mtDupKey); found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"retcode": "E0023",
			"message": "Duplicate merchant_transaction_id",
		})
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

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)
	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	isBlocked, _ := repository.IsUserIDBlocked(transaction.UserId, arrClient.ClientName)
	if isBlocked {
		log.Println("userID is blocked")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "userID is blocked",
		})
	}

	appName := repository.GetAppNameFromClient(arrClient, appid)

	expectedBodysign := helper.GenerateBodySign(transaction, arrClient.ClientSecret)
	// log.Println("expectedBodysign", expectedBodysign)

	if receivedBodysign != expectedBodysign {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid bodysign",
		})
	}

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

	createdTransId, chargingPrice, err := repository.CreateTransaction(context.Background(), &transaction, arrClient, appkey, appid, &vaNumber)
	if err != nil {
		log.Println("err", err)
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	MTIDCache.Set(mtDupKey, true, cache.DefaultExpiration)

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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success":      true,
			"retcode":      "0000",
			"phone_number": transaction.UserMDN,
			"keyword":      keyword,
			"sms_code":     fmt.Sprintf("%d", otp),
			"short_code":   "99899",
			"trx_type":     "send_otp",
			"message":      "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success":      true,
			"reference_id": ximpayId,
			"guide": fiber.Map{
				"en": "Please enter the OTP received via SMS",
				"id": "Mohon masukan otp yang diterima di sms",
			},
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		// log.Println("redirect: ", res.Actions[0].URL)
		return c.JSON(fiber.Map{
			"success":  true,
			"redirect": res.Actions[0].URL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
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

		// log.Println("redirect: ", res.Actions[0].URL)
		return c.JSON(fiber.Map{
			"success":  true,
			"redirect": res.Actions[1].URL,
			"qrisUrl":  res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
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

		// log.Println("redirect: ", res.Actions[0].URL)
		return c.JSON(fiber.Map{
			"success":  true,
			"qrisUrl":  res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"qrString": res.QrString,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
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
		// res, err := lib.RequestChargingDanaFaspay(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL, transaction.CustomerName, transaction.UserMDN) //lib.RequestChargingDana(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL)
		checkoutUrl, err := lib.RequestChargingDana(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request dana failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		// for faspay
		// if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1001, &res.TrxID, nil, "", nil); err != nil {
		// 	log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
		// }

		return c.JSON(fiber.Map{
			"success":  true,
			"back_url": transaction.RedirectURL,
			"redirect": checkoutUrl, //"redirect": res.RedirectURL,
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

func CreateTransactionV1(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CreateTransactionV1", "handler")
	defer span.End()

	bodysign := c.Get("bodysign")
	appkey := c.Get("appkey")
	appid := c.Get("appid")

	var transaction model.InputPaymentRequest
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
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

	if transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN , item_name or amounr must not be empty",
		})
	}

	beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	isBlocked, err := repository.IsMDNBlocked(beautifyMsisdn)
	if err != nil {
		log.Println("Msisdn is blocked")

	}

	if isBlocked {
		log.Println(" diblokir")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Msisdn is blocked",
		})
	}

	if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Phone number %s is inactive or invalid, please try another number", transaction.UserMDN),
		})

	}

	if !helper.IsValidPrefix(beautifyMsisdn, paymentMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid prefix, please use valid phone number.",
		})
	}

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)
	if err != nil {
		log.Println("Error get client")

	}

	isBlocked, _ = repository.IsUserIDBlocked(transaction.UserId, arrClient.ClientName)
	if isBlocked {
		log.Println("userID is blocked")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "userID is blocked",
		})

	}

	appName := repository.GetAppNameFromClient(arrClient, appid)

	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	transaction.BodySign = bodysign
	arrClient.AppName = appName
	transaction.PaymentMethod = paymentMethod

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	createdTransId, chargingPrice, err := repository.CreateTransaction(spanCtx, &transaction, arrClient, appkey, appid, nil)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

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
				"message": fmt.Sprintf("request failed: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
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
			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		err = repository.UpdateXimpayID(context.Background(), createdTransId, ximpayId)
		if err != nil {
			log.Println("Updated Ximpay ID error:", err)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"reff_id": ximpayId,
			"message": "Successful Created Transaction",
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

		return c.JSON(fiber.Map{
			"success": true,
			"retcode": "0000",
			"message": "Successful Created Transaction",
		})

	}

	return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"retcode": "0000",
		"message": "Successful Created Transaction",
	})
}

func CreateTransactionNonTelco(c *fiber.Ctx) error {
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

	var isMidtrans bool

	var paymentMethod string
	switch transaction.PaymentMethod {
	case "ovo_wallet":
		paymentMethod = "ovo"
	case "qr":
		paymentMethod = "qris"
	default:
		paymentMethod = transaction.PaymentMethod

	}

	if paymentMethod == "shopeepay" || paymentMethod == "gopay" || paymentMethod == "qris" || paymentMethod == "dana" {
		isMidtrans = true
	}

	if !isMidtrans && (transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN , item_name or amount must not be empty",
		})
	}

	// beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	// isBlocked, err := repository.IsMDNBlocked(beautifyMsisdn)
	// if err != nil {
	// 	log.Println("Msisdn is blocked")

	// }

	// if isBlocked {
	// 	log.Println(" diblokir")
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Msisdn is blocked",
	// 	})
	// }

	// if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": fmt.Sprintf("Phone number %s is inactive or invalid, please try another number", transaction.UserMDN),
	// 	})

	// }

	// if !helper.IsValidPrefix(beautifyMsisdn, transaction.PaymentMethod) {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"success": false,
	// 		"error":   "Invalid prefix, please use valid phone number.",
	// 	})
	// }

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)

	appName := repository.GetAppNameFromClient(arrClient, appid)

	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	transaction.BodySign = bodysign
	arrClient.AppName = appName

	createdTransId, chargingPrice, err := repository.CreateTransaction(spanCtx, &transaction, arrClient, appkey, appid, nil)
	if err != nil {
		log.Println("err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Internal Server Error", "response": "error create transaction", "data": err})
	}

	switch paymentMethod {
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
		TransactionCache.Delete(token)
		return c.JSON(fiber.Map{
			"success":  true,
			"redirect": res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
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

		TransactionCache.Delete(token)
		return c.JSON(fiber.Map{
			"success":  true,
			"redirect": res.Actions[1].URL,
			"qrisUrl":  res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
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

		TransactionCache.Delete(token)
		return c.JSON(fiber.Map{
			"success":  true,
			"qrisUrl":  res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
		})
	case "ovo":
		res, err := lib.ChargingOVO(createdTransId, chargingPrice, transaction.UserMDN)
		if err != nil {
			log.Println("Charging request ovo failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		referenceId := fmt.Sprintf("%s-%s", res.ApprovalCode, res.TransactionRequestData.MerchantInvoice)

		now := time.Now()

		receiveCallbackDate := &now

		switch res.ResponseCode {
		case "00":
			log.Println("res charge ovo:", res)
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1003, &referenceId, nil, "", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "13":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Invalid amount", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "14":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Invalid Mobile Number / OVO ID", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "17":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Decline", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "25":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Not Found", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "26":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Failed", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "40":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Failed", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "68":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Transaction Pending / Timeout", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "94":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Duplicate Request Params", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "ER":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "System Failure", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "EB":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Terminal Blocked", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		case "BR":
			if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1005, &referenceId, nil, "Bad Request", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
			}
		}

		TransactionCache.Delete(token)
		return c.JSON(fiber.Map{
			"success": true,
			// "qrisUrl":  res.Actions[0].URL,
			"back_url": transaction.RedirectURL,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
		})
	case "dana":
		strPrice := fmt.Sprintf("%d00", chargingPrice)
		// res, err := lib.RequestChargingDanaFaspay(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL, transaction.CustomerName, transaction.UserMDN) //lib.RequestChargingDana(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL)
		// if err != nil {
		// 	log.Println("Charging request dana failed:", err)
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"success": false,
		// 		"message": "Charging request failed",
		// 	})
		// }

		checkoutUrl, err := lib.RequestChargingDana(createdTransId, transaction.ItemName, strPrice, transaction.RedirectURL)
		if err != nil {
			log.Println("Charging request dana failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		// if err := repository.UpdateTransactionStatus(context.Background(), createdTransId, 1001, &res.TrxID, nil, "", nil); err != nil {
		// 	log.Printf("Error updating transaction status for %s: %s", createdTransId, err)
		// }

		// err = repository.UpdateMidtransId(context.Background(), createdTransId, res.TransactionID)
		// if err != nil {
		// 	log.Println("Updated Midtrans ID error:", err)
		// }

		TransactionCache.Delete(token)
		return c.JSON(fiber.Map{
			"success":  true,
			"back_url": transaction.RedirectURL,
			"redirect": checkoutUrl,
			"retcode":  "0000",
			"message":  "Successful Created Transaction",
		})
	}

	return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"retcode": "0000",
		"message": "Successful Created Transaction",
	})
}

func GetTransactions(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "GetTransactions", "handler")
	defer span.End()

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	appID := c.Query("app_id")
	userMDN := helper.BeautifyIDNumber(c.Query("user_mdn"), true)
	paymentMethodStr := c.Query("payment_method")
	var paymentMethods []string
	if paymentMethodStr != "" {
		paymentMethods = strings.Split(paymentMethodStr, ",")
	} else {
		paymentMethods = []string{}
	}
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userId := c.Query("user_id")
	transactionId := c.Query("transaction_id")
	merchantTransactionId := c.Query("merchant_transaction_id")
	appName := c.Query("app_name")
	merchantNameStr := c.Query("merchant_name")
	var merchants []string
	if merchantNameStr != "" {
		merchants = strings.Split(merchantNameStr, ",")
	} else {
		merchants = []string{}
	}
	denomStr := c.Query("denom")
	denom, err := strconv.Atoi(denomStr)
	if err != nil {
		fmt.Println("error convert status")
	}
	statusStr := c.Query("status")

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		fmt.Println("error convert status")
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStartDate
		}
	}
	if endDateStr != "" {
		parsedEndDate, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEndDate
		}
	}

	// excludeMerchantStr := c.Query("exclude_merchant")
	// var excludeMerchants []string
	// if excludeMerchantStr != "" {
	// 	excludeMerchants = strings.Split(excludeMerchantStr, ",")
	// }

	transactions, totalItems, err := repository.GetAllTransactions(spanCtx, limit, offset, status, denom, transactionId, merchantTransactionId, appID, userMDN, userId, appName, merchants, paymentMethods, startDate, endDate)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(limit)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transactions,
		"pagination": fiber.Map{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    totalItems,
			"items_per_page": limit,
		},
	})
}

func ExportTransactions(c *fiber.Ctx) error {
	_, spanCtx := apm.StartSpan(c.Context(), "ExportTransactionsCSV", "handler")
	exportCSV := c.Query("export_csv", "false")
	exportExcel := c.Query("export_excel", "false")

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	statusStr := c.Query("status")
	paymentMethod := c.Query("payment_method")
	appName := c.Query("app_name")
	merchantName := c.Query("merchant_name")

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		fmt.Println("error convert status")
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStartDate
		}
	}
	if endDateStr != "" {
		parsedEndDate, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEndDate
		}
	}

	transactions, err := repository.GetTransactionsByDateRange(spanCtx, status, startDate, endDate, merchantName, appName, paymentMethod)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	if exportCSV == "true" {
		return exportTransactionsToCSV(c, transactions)
	}

	if exportExcel == "true" {
		return exportTransactionsToExcel(c, transactions)
	}

	return nil

}

func ExportTransactionsMerchant(c *fiber.Ctx) error {
	_, spanCtx := apm.StartSpan(c.Context(), "ExportTransactionsCSV", "handler")
	exportCSV := c.Query("export_csv", "false")
	exportExcel := c.Query("export_excel", "false")

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	paymentMethod := c.Query("paymentMethod")

	statusStr := c.Query("status")
	appKey := c.Get("appkey")
	appID := c.Get("appid")

	if appKey == "" || appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing header: appkey & appid",
		})
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		fmt.Println("error convert status")
	}

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStartDate
		}
	}
	if endDateStr != "" {
		parsedEndDate, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEndDate
		}
	}

	arrClient, err := repository.FindClient(context.Background(), appKey, appID)
	if err != nil {
		fmt.Println("Error fetching client:", err)
	}

	transactions, err := repository.GetTransactionsByDateRange(spanCtx, status, startDate, endDate, arrClient.ClientName, arrClient.AppName, paymentMethod)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	if exportCSV == "true" {
		return exportTransactionsToCSV(c, transactions)
	}

	if exportExcel == "true" {
		return exportTransactionsToExcel(c, transactions)
	}

	return nil

}

func exportTransactionsToCSV(c *fiber.Ctx, transactions []model.Transactions) error {
	log.Println("export transaction csv hit")
	// Set header untuk file CSV
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=transactions.csv")

	if len(transactions) > 200000 {
		return response.Response(c, fiber.StatusBadRequest, "Data terlalu besar untuk diekspor ke Excel. Silakan gunakan CSV.")
	}

	// Buat writer untuk CSV
	writer := csv.NewWriter(c)
	defer writer.Flush()

	header := []string{"ID", "Merchant Transaction ID", "Date", "MDN", "Merchant", "App", "Amount", "Price", "Fee", "Item", "Method", "Net Amount", "User ID", "Currency", "Item ID", "Status"}
	if err := writer.Write(header); err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	for _, transaction := range transactions {
		var status string
		var price uint
		var fee uint
		var netAmount uint
		switch transaction.StatusCode {
		case 1005:
			status = "failed"
		case 1001:
			status = "pending"
		case 1003:
			status = "pending"
		case 1000:
			status = "success"
		}

		switch transaction.PaymentMethod {
		case "qris":
			feeFloat := float64(transaction.Amount) * 0.008
			fee = uint(math.Ceil(feeFloat))
			price = transaction.Amount
			netAmount = price - fee
		case "dana":
			feeFloat := float64(transaction.Amount) * 0.018
			fee = uint(math.Ceil(feeFloat))
			price = transaction.Amount
			netAmount = price - fee
		default:
			price = transaction.Price
		}

		var createdAt string
		if transaction.AppName == "Zingplay games" {
			createdAt = transaction.CreatedAt.In(loc).Format("01/02/2006 15:04:05")
		} else {
			createdAt = transaction.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		}

		record := []string{
			transaction.ID,
			transaction.MtTid,
			createdAt,
			transaction.UserMDN,
			transaction.MerchantName,
			transaction.AppName,
			strconv.Itoa(int(transaction.Amount)),
			strconv.Itoa(int(price)),
			strconv.Itoa(int(fee)),
			transaction.ItemName,
			transaction.PaymentMethod,
			strconv.Itoa(int(netAmount)),
			transaction.UserId,
			transaction.Currency,
			transaction.ItemId,
			status,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func exportTransactionsToExcel(c *fiber.Ctx, transactions []model.Transactions) error {

	log.Println("export transaction excel hit")
	f := excelize.NewFile()
	sheetName := "Transactions"
	index, _ := f.NewSheet(sheetName)

	if len(transactions) > 80000 {
		return response.Response(c, fiber.StatusBadRequest, "Data terlalu besar untuk diekspor ke Excel. Silakan gunakan CSV.")
	}

	// Tulis header
	headers := []string{"ID", "Merchant Transaction ID", "Date", "MDN", "Merchant", "App", "Amount", "Price", "Fee", "Item", "Method", "Net Amount", "User ID", "Currency", "Item ID", "Status"}
	for i, header := range headers {
		cell := getColumnName(i+1) + "1"
		f.SetCellValue(sheetName, cell, header)
		// f.SetCellStyle(sheetName, cell, cell, `{"font":{"bold":true}}`) // Set header menjadi bold
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	// Tulis data transaksi
	for rowIndex, transaction := range transactions {
		var status string
		var price uint
		var fee uint
		var netAmount uint
		switch transaction.StatusCode {
		case 1005:
			status = "failed"
		case 1001:
			status = "pending"
		case 1003:
			status = "pending"
		case 1000:
			status = "success"
		}

		switch transaction.PaymentMethod {
		case "qris":
			price = transaction.Amount
			feeFloat := float64(transaction.Amount) * 0.008
			fee = uint(math.Ceil(feeFloat))
			netAmount = price - fee
		case "dana":
			feeFloat := float64(transaction.Amount) * 0.018
			fee = uint(math.Ceil(feeFloat))
			price = transaction.Amount
			netAmount = price - fee
		default:
			price = transaction.Price
		}

		var createdAt string

		if transaction.AppName == "Zingplay games" {
			createdAt = transaction.CreatedAt.In(loc).Format("01/02/2006 15:04:05")
		} else {
			createdAt = transaction.CreatedAt.In(loc).Format("2006-01-02 15:04:05")
		}

		row := rowIndex + 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), transaction.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), transaction.MtTid)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), createdAt)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), transaction.UserMDN)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), transaction.MerchantName)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), transaction.AppName)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), transaction.Amount)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), price)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(row), fee)
		f.SetCellValue(sheetName, "J"+strconv.Itoa(row), transaction.ItemName)
		f.SetCellValue(sheetName, "K"+strconv.Itoa(row), transaction.PaymentMethod)
		f.SetCellValue(sheetName, "L"+strconv.Itoa(row), netAmount)
		f.SetCellValue(sheetName, "M"+strconv.Itoa(row), transaction.UserId)
		f.SetCellValue(sheetName, "N"+strconv.Itoa(row), transaction.Currency)
		f.SetCellValue(sheetName, "O"+strconv.Itoa(row), transaction.ItemId)
		f.SetCellValue(sheetName, "P"+strconv.Itoa(row), status)
	}

	for i := 0; i < len(headers); i++ {
		// cell := getColumnName(i+1) + "1"
		// f.SetCellStyle(sheetName, cell, cell,)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Simpan file Excel
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=transactions.xlsx")

	return f.Write(c)
}

// Fungsi untuk mendapatkan nama kolom berdasarkan indeks
func getColumnName(index int) string {
	columnName := ""
	for index > 0 {
		index-- // Mengurangi 1 untuk mengubah indeks ke 0-based
		columnName = string('A'+(index%26)) + columnName
		index /= 26
	}
	return columnName
}

func GetTransactionByID(c *fiber.Ctx) error {
	id := c.Params("id")

	transaction, err := repository.GetTransactionByID(context.Background(), id)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transaction,
	})
}

func GetTransactionMerchantByID(c *fiber.Ctx) error {
	id := c.Params("id")
	appKey := c.Get("appkey")
	appID := c.Get("appid")

	transaction, err := repository.GetTransactionMerchantByID(context.Background(), appKey, appID, id)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transaction,
	})
}

func GetTransactionsMerchant(c *fiber.Ctx) error {
	appKey := c.Get("appkey")
	appID := c.Get("appid")

	if appKey == "" || appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing header: appkey & appid",
		})
	}

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	exportCSV := c.Query("export_csv", "false")
	exportExcel := c.Query("export_excel", "false")
	userId := c.Query("user_id")

	userMDN := c.Query("user_mdn")
	paymentMethodStr := c.Query("payment_method")
	var paymentMethods []string
	if paymentMethodStr != "" {
		paymentMethods = strings.Split(paymentMethodStr, ",")
	} else {
		paymentMethods = []string{}
	}
	merchantTransactionId := c.Query("merchant_transaction_id")
	appName := c.Query("app_name")
	denomStr := c.Query("denom")
	denom, err := strconv.Atoi(denomStr)
	if err != nil {
		fmt.Println("error convert status")
	}
	statusStr := c.Query("status")

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		fmt.Println("error convert status")
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsedStartDate, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStartDate
		}
	}
	if endDateStr != "" {
		parsedEndDate, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEndDate
		}
	}

	arrClient, err := repository.FindClient(context.Background(), appKey, appID)
	if err != nil {
		fmt.Println("Error fetching client:", err)
	}

	transactions, totalItems, err := repository.GetTransactionsMerchant(context.Background(), limit, offset, status, denom, merchantTransactionId, arrClient.ClientName, userMDN, userId, appName, paymentMethods, startDate, endDate)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	if exportCSV == "true" {
		return exportTransactionsToCSV(c, convertToExportFormat(transactions))
	}

	if exportExcel == "true" {
		return exportTransactionsToExcel(c, convertToExportFormat(transactions))
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(limit)))

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transactions,
		"pagination": fiber.Map{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    totalItems,
			"items_per_page": limit,
		},
	})

}

func convertToExportFormat(transactions []model.TransactionMerchantResponse) []model.Transactions {
	var exportData []model.Transactions
	for _, transaction := range transactions {
		exportData = append(exportData, model.Transactions{
			ID:                      transaction.ID,
			UserMDN:                 transaction.UserMDN,
			UserId:                  transaction.UserID,
			PaymentMethod:           transaction.PaymentMethod,
			MtTid:                   transaction.MerchantTransactionID,
			AppName:                 transaction.AppName,
			StatusCode:              transaction.StatusCode,
			TimestampRequestDate:    transaction.TimestampRequestDate,
			TimestampSubmitDate:     transaction.TimestampSubmitDate,
			TimestampCallbackDate:   transaction.TimestampCallbackDate,
			TimestampCallbackResult: transaction.TimestampCallbackResult,
			ItemId:                  transaction.ItemId,
			ItemName:                transaction.ItemName,
			Route:                   transaction.Route,
			Currency:                transaction.Currency,
			Amount:                  transaction.Amount,
			Price:                   transaction.Price,
			CreatedAt:               transaction.CreatedAt,
			UpdatedAt:               transaction.UpdatedAt,
		})
	}
	return exportData
}

func ManualCallback(c *fiber.Ctx) error {
	transactionID := c.Params("id")

	transaction, err := repository.GetTransactionByID(c.Context(), transactionID)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, "Transaction not found")
	}

	if transaction.StatusCode != 1000 && transaction.StatusCode != 1003 {
		return response.Response(c, fiber.StatusBadRequest, "Transaction not success")
	}
	arrClient, err := repository.FindClient(context.Background(), transaction.ClientAppKey, transaction.AppID)
	if err != nil {
		fmt.Println("Error fetching client:", err)
	}

	var callbackURL string
	for _, app := range arrClient.ClientApps {
		if app.AppID == transaction.AppID {
			callbackURL = app.CallbackURL
			break
		}
	}

	if transaction.NotificationUrl != "" {
		callbackURL = transaction.NotificationUrl
	}

	if callbackURL == "" {
		log.Printf("No matching ClientApp found for AppID callback Url: %s", transaction.AppID)
	}

	statusCode := 1000

	var paymentMethod string

	paymentMethod = transaction.PaymentMethod
	if transaction.MerchantName == "HIGO GAME PTE LTD" && transaction.PaymentMethod == "qris" {
		paymentMethod = "qr"
	}

	var amount interface{}
	if arrClient.ClientName == "WEIDIAN TECHNOLOGY CO" || arrClient.ClientSecret == "o_G0JIzzJLditvj" {
		amount = transaction.Amount
	} else {
		amount = fmt.Sprintf("%d", transaction.Amount)
	}

	// callbackData := repository.CallbackData{
	// 	UserID:                transaction.UserId,
	// 	MerchantTransactionID: transaction.MtTid,
	// 	StatusCode:            statusCode,
	// 	PaymentMethod:         paymentMethod,
	// 	Amount:                amount,
	// 	Status:                "success",
	// 	Currency:              transaction.Currency,
	// 	ItemName:              transaction.ItemName,
	// 	ItemID:                transaction.ItemId,
	// 	ReferenceID:           transactionID,
	// }

	// if arrClient.ClientName == "Zingplay International PTE,. LTD" || arrClient.ClientSecret == "9qyxr81YWU2BNlO" {
	// 	callbackData.AppID = transaction.AppID
	// 	callbackData.ClientAppKey = transaction.ClientAppKey
	// }

	var callbackPayload interface{}

	if arrClient.ClientName == "PM Max" || arrClient.ClientSecret == "gmtb50vcf5qcvwr" ||
		arrClient.ClientName == "Coda" || arrClient.ClientSecret == "71mczdtiyfaunj5" {

		callbackPayload = model.CallbackDataLegacy{
			AppID:                  transaction.AppID,
			ClientAppKey:           transaction.ClientAppKey,
			UserID:                 transaction.UserId,
			UserIP:                 transaction.UserIP,
			UserMDN:                transaction.UserMDN,
			MerchantTransactionID:  transaction.MtTid,
			TransactionDescription: "",
			PaymentMethod:          paymentMethod,
			Currency:               transaction.Currency,
			Amount:                 transaction.Amount,
			ChargingAmount:         fmt.Sprintf("%d", transaction.Price),
			StatusCode:             fmt.Sprintf("%d", statusCode),
			Status:                 "success",
			ItemID:                 transaction.ItemId,
			ItemName:               transaction.ItemName,
			UpdatedAt:              fmt.Sprintf("%d", time.Now().Unix()),
			ReferenceID:            transaction.CallbackReferenceId,
			Testing:                "0",
			Custom:                 "",
		}
	} else {
		payload := repository.CallbackData{
			UserID:                transaction.UserId,
			MerchantTransactionID: transaction.MtTid,
			StatusCode:            statusCode,
			PaymentMethod:         paymentMethod,
			Amount:                amount,
			Status:                "success",
			Currency:              transaction.Currency,
			ItemName:              transaction.ItemName,
			ItemID:                transaction.ItemId,
			ReferenceID:           transaction.ID,
		}

		if arrClient.ClientName == "Zingplay International PTE,. LTD" || arrClient.ClientSecret == "9qyxr81YWU2BNlO" {
			payload.AppID = transaction.AppID
			payload.ClientAppKey = transaction.ClientAppKey
		}

		callbackPayload = payload
	}

	err = repository.SendCallback(callbackURL, arrClient.ClientSecret, transaction.ID, callbackPayload)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"success": true, "message": "Callback sent successfully"})
}

func CheckTrans(c *fiber.Ctx) error {
	id := c.Params("id")

	config, _ := config.GetGatewayConfig("xl_twt")
	arrayOptions := config.Options["development"].(map[string]interface{})

	token, _ := lib.GetAccessTokenXl(arrayOptions["clientid"].(string), arrayOptions["clientsecret"].(string))

	status, err := lib.CheckTransactions(id, "RDSN", token)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    status,
	})
}

func MTSmartfren(c *fiber.Ctx) error {

	var transaction model.InputPaymentRequest
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	if transaction.ReffId == "" || transaction.Otp == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: ReffId, Otp must not be empty",
		})
	}

	err := lib.DoMT(transaction.ReffId, transaction.Otp)
	if err != nil {
		log.Println("MT request smartfren failed:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Charging request failed",
		})
	}

	return response.ResponseSuccess(c, fiber.StatusOK, fiber.Map{
		"success": true,
		"message": "Successful Transaction",
	})
}
