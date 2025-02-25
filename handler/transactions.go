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

func CheckedTransaction(paymentRequest *http.CreatePaymentRequest, client *model.Client) map[string]interface{} {
	var chargingPrice float64

	// fmt.Printf("Client details: %+v\n", client)

	if len(paymentRequest.UserID) > 50 || len(paymentRequest.MerchantTransactionID) > 36 || len(paymentRequest.ItemName) > 25 {
		fmt.Printf("Too long parameters: %+v\n", paymentRequest)
		return map[string]interface{}{
			"success": false,
			"retcode": "E0021",
		}
	}

	paymentMethod := paymentRequest.PaymentMethod
	switch paymentMethod {
	case "xl_gcpay", "xl_gcpay2":
		paymentMethod = "xl_airtime"
	case "smartfren":
		paymentMethod = "smartfren_airtime"
	case "three":
		paymentMethod = "three_airtime"
	case "telkomsel_airtime_sms", "telkomsel_airtime_ussd", "telkomsel_airtime_mdm":
		paymentMethod = "telkomsel_airtime"
	case "indosat_huawei", "indosat_mimopay", "indosat_simplepayment":
		paymentMethod = "indosat_airtime"
	}
	// log.Println("test log")

	arrPaymentMethod, err := repository.FindPaymentMethodBySlug(paymentMethod, "")
	if err != nil {
		log.Println(err.Error())
		return map[string]interface{}{
			"message": err.Error(),
			"success": false,
			"retcode": "E0005",
		}
	}

	var route string

	// Loop through client's payment methods

	// route di paymentmethod client bisa status int atau  array denom -aldi

	// log.Printf("payment: %+v\n", arrPaymentMethod)

	// Checking the payment method on the client is matched with the payment_method from the request. If it is not flexible, take a route in the payment method that contains the amount. If it is flexible, take a route that has a value of 1.
	for _, arrayPayments := range client.PaymentMethods {

		if arrayPayments.Name == paymentMethod {

			arrRoutes := arrayPayments.Route
			if !arrPaymentMethod.Flexible {

				// for routename, arrayDenom := range arrRoutes {
				// 	denom := arrayDenom
				// 	// Check if the amount is in the denominated range
				// 	log.Printf("routeName: %v    ", denom)

				// 	if strSlice, ok := denom.([]string); ok && containsString(strSlice, fmt.Sprintf("%.0f", paymentRequest.Amount)) {

				// 		route = routename
				// 		break
				// 	} else {
				// 		fmt.Printf("Invalid type for arrayDenom: %T\n", arrayDenom)
				// 	}
				// }
				// Metode pembayaran non-fleksibel
				for routename, arrayDenom := range arrRoutes {

					log.Println("routename and arrDenom: ", routename, arrayDenom)
					// Cek apakah arrayDenom adalah slice interface{}

					// if denomSlice, ok := arrayDenom.(primitive.A); ok {
					// 	// Konversi ke slice string
					// 	stringSlice := make([]string, len(denomSlice))
					// 	for i, v := range denomSlice {
					// 		stringSlice[i] = fmt.Sprintf("%v", v) // Mengkonversi setiap elemen ke string
					// 	}

					// 	// Cek apakah amount ada dalam denominasi
					// 	if containsString(stringSlice, fmt.Sprintf("%.0f", paymentRequest.Amount)) {
					// 		route = routename
					// 		break
					// 	}
					// } else {
					// 	fmt.Printf("Invalid type for arrayDenom: %T\n", arrayDenom)
					// }
				}

				// for routename, arrayDenom := range arrRoutes {

				// 	if denom, exists := arrRoutes["xl_twt"]; exists {
				// 		if denomSlice, ok := arrayDenom.(primitive.A); ok {
				// 			// Konversi ke slice string
				// 			stringSlice := make([]string, len(denomSlice))
				// 			for i, v := range denomSlice {
				// 				stringSlice[i] = fmt.Sprintf("%v", v) // Mengkonversi setiap elemen ke string
				// 			}

				// 			// Cek apakah amount ada dalam denominasi
				// 			if contains(stringSlice, fmt.Sprintf("%.0f", paymentRequest.Amount)) {
				// 				route = routename
				// 				// log.Println(route)
				// 				// log.Println("Using non-flexible route.")
				// 			}
				// 		} else {
				// 			log.Printf("Invalid type for xl_twt: %T\n", denom)
				// 		}
				// 	}

				// }
			} else {
				// for routename, value := range arrRoutes {
				// 	if valueStr, ok := value.(string); ok && valueStr == "1" {
				// 		route = routename
				// 		break
				// 	}
				// }
			}
		}
	}

	if route == "" {
		return map[string]interface{}{
			"success": false,
			"retcode": "E0007",
		}
	}

	// 	// TODO
	// 	// perlu checksupported di repository, check code legacy
	// 	// check func search_sub_array di code legacy

	// 	// TODO
	// 	// The logic for define the charging price is not yet complete
	arrPaymentMethodRoute, err := repository.FindPaymentMethodBySlug(route, "")

	if arrPaymentMethodRoute.Flexible {
		if paymentRequest.Amount < arrPaymentMethodRoute.MinimumDenom {
			if client.Testing == 0 {
				return map[string]interface{}{
					"success": false,
					"retcode": "E0020",
				}
			}
			chargingPrice = float64(paymentRequest.Amount)
		} else {
			switch {
			case paymentMethod == "indosat_airtime2" && route == "indosat_triyakom4":
				chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
			case paymentMethod == "gopay" && route == "gopay_midtrans":
				clientName := client.ClientName
				if clientName == "Topfun 2 New Qiuqiu" || clientName == "Topfun" || clientName == "SPOLIVE" || clientName == "Tricklet (Hong Kong) Limited" {
					chargingPrice = chargingPrice
				} else {
					chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
				}
			case paymentMethod == "shopeepay" && route == "shopeepay_midtrans":
				clientName := client.ClientName
				if clientName == "Tricklet (Hong Kong) Limited" {
					chargingPrice = chargingPrice
				} else {
					chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
				}
			case paymentMethod == "alfamart_otc" && route == "alfamart_faspay":
				clientName := client.ClientName
				if clientName == "Tricklet (Hong Kong) Limited" {
					chargingPrice = chargingPrice
				} else if clientName == "Redigame" {
					chargingPrice = chargingPrice + 6000
				} else {
					chargingPrice = chargingPrice + 6500
				}
			case paymentMethod == "indomaret_otc" && route == "indomaret_otc_mst":
				clientName := client.ClientName
				subtotal := chargingPrice
				var adminFee, totalAmount float64
				switch clientName {
				case "Tricklet (Hong Kong) Limited":
					adminFee = 0
				case "Redigame":
					adminFee = ((0.06 * chargingPrice) + 1000) / (1 - 0.06)
				case "Higo Game PTE LTD":
					adminFee = ((0.07 * chargingPrice) + 1000) / (1 - 0.07)
				default:
					adminFee = ((0.075 * chargingPrice) + 1000) / (1 - 0.075)
				}
				totalAmount = subtotal + adminFee
				chargingPrice = math.Round(totalAmount/100) * 100
			case paymentMethod == "smartfren_airtime2" && (route == "smartfren_triyakom_flex" || route == "smartfren_triyakom_flex2"):
				chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
			case paymentMethod == "three_airtime2" && route == "three_triyakom_flex2":
				chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
			default:
				chargingPrice = chargingPrice
			}

			if client.ClientName == "Zingplay International PTE,. LTD" && paymentMethod == "ovo_wallet" && route == "ovo" {
				chargingPrice = chargingPrice + math.Round(0.11*chargingPrice)
			}
		}

	} else {

		// TODO
		// recheck jika perlu pengecekan denom saat request charging, bisa check code legacy

		if arrPaymentMethodRoute.IsAirtime == "1" {
			// Check length of MDN
			if len(helper.BeautifyIDNumber(paymentRequest.UserMDN, false)) > 14 {
				return map[string]interface{}{
					"success": false,
					"retcode": "E0016",
				}
			}

			_, err := helper.ByPrefixNumber(route, helper.BeautifyIDNumber(paymentRequest.UserMDN, false))
			if err != nil {
				return map[string]interface{}{
					"success": false,
					"retcode": "E0016",
				}
			}
		}

		price, _ := repository.GetPrice(route, paymentRequest.Amount)
		chargingPrice = float64(price)
	}
	chargingPrice = float64(paymentRequest.Amount) + math.Round(0.11*float64(paymentRequest.Amount))

	return map[string]interface{}{
		"success":        true,
		"retcode":        "E0020",
		"mobile":         client.Mobile,
		"testing":        client.Testing,
		"charging_price": float32(chargingPrice),
		"route":          route,
		"payment_method": paymentMethod,
	}
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

	var transaction model.InputPaymentRequest
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	if transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN, item_name or amount must not be empty",
		})
	}

	beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Phone number %s is inactive or invalid, please try another number", transaction.UserMDN),
		})

	}

	if !helper.IsValidPrefix(beautifyMsisdn, transaction.PaymentMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid prefix, please use valid phone number.",
		})
	}

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)
	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	appName := repository.GetAppNameFromClient(arrClient, appid)

	// expectedBodysign := helper.GenerateBodySign(transaction, arrClient.ClientSecret)
	// log.Println("arrClient", arrClient)

	paymentMethodMap := make(map[string]model.PaymentMethodClient)
	for _, pm := range arrClient.PaymentMethods {
		paymentMethodMap[pm.Name] = pm
	}

	paymentMethodClient, exists := paymentMethodMap[transaction.PaymentMethod]
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

	createdTransId, chargingPrice, err := repository.CreateTransaction(context.Background(), &transaction, arrClient, appkey, appid)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	// createdTransId := "rawgg36"
	// var chargingPrice uint
	// chargingPrice = 5550

	switch transaction.PaymentMethod {
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
	case "tri_airtime":
		validAmounts, exists := routes["tri_triyakom"]
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

		// validAmounts, exists := routes["indosat_triyakom"]
		// if !exists {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"error": "No valid amounts found for the specified payment method",
		// 	})
		// }

		// validAmount := false
		// for _, route := range validAmounts {
		// 	if transactionAmountStr == route {
		// 		validAmount = true
		// 		break
		// 	}
		// }

		// if !validAmount && !paymentMethodClient.Flexible {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"error": "This denom is not supported for this payment method",
		// 	})
		// }

		_, keyword, otp, err := lib.RequestMoTsel(transaction.UserMDN, transaction.MtTid, transaction.ItemName, createdTransId, transactionAmountStr)
		if err != nil {
			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		err = repository.UpdateTransactionKeyword(context.Background(), createdTransId, keyword, otp)
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

	if transaction.UserId == "" || transaction.MtTid == "" || transaction.UserMDN == "" || transaction.PaymentMethod == "" || transaction.Amount <= 0 || transaction.ItemName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing mandatory fields: UserId, mtId, paymentMethod, UserMDN , item_name or amounr must not be empty",
		})
	}
	beautifyMsisdn := helper.BeautifyIDNumber(transaction.UserMDN, false)

	if _, found := lib.NumberCache.Get(beautifyMsisdn); found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Phone number %s is inactive or invalid, please try another number", transaction.UserMDN),
		})

	}

	if !helper.IsValidPrefix(beautifyMsisdn, transaction.PaymentMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid prefix, please use valid phone number.",
		})
	}

	arrClient, err := repository.FindClient(spanCtx, appkey, appid)

	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	transaction.BodySign = bodysign

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	createdTransId, chargingPrice, err := repository.CreateTransaction(spanCtx, &transaction, arrClient, appkey, appid)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	paymentMethodMap := make(map[string]model.PaymentMethodClient)
	for _, pm := range arrClient.PaymentMethods {
		paymentMethodMap[pm.Name] = pm
	}

	paymentMethodClient, exists := paymentMethodMap[transaction.PaymentMethod]
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

	switch transaction.PaymentMethod {
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

		if !validAmount {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This denom is not supported for this payment method",
			})
		}
	case "tri_airtime":
		validAmounts, exists := routes["tri_triyakom"]
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

	transactions, err := repository.GetTransactionsByDateRange(spanCtx, startDate, endDate)
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
	// Set header untuk file CSV
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=transactions.csv")

	// Buat writer untuk CSV
	writer := csv.NewWriter(c)
	defer writer.Flush()

	// Tulis header CSV
	header := []string{"ID", "Date", "MT TID", "Payment Method", "Amount", "User ID", "App Name", "Currency", "Item Name", "Item ID", "Status Code"}
	if err := writer.Write(header); err != nil {
		return err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")

	for _, transaction := range transactions {
		var status string
		switch transaction.StatusCode {
		case 1005:
			status = "Failed"
		case 1001 | 1003:
			status = "Pending"
		case 1000:
			status = "Success"
		}

		createdAt := transaction.CreatedAt.In(loc).Format("2006-01-02 15:04:05")

		record := []string{
			transaction.ID,
			createdAt,
			transaction.MtTid,
			transaction.PaymentMethod,
			strconv.Itoa(int(transaction.Amount)),
			transaction.UserId,
			transaction.AppName,
			transaction.Currency,
			transaction.ItemName,
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
	// Buat file Excel baru
	f := excelize.NewFile()
	sheetName := "Transactions"
	index, _ := f.NewSheet(sheetName)

	// Tulis header
	headers := []string{"Transaction ID", "Date", "MT TID", "Payment Method", "Amount", "User ID", "App Name", "Currency", "Item Name", "Item ID", "Status Code"}
	for i, header := range headers {
		cell := getColumnName(i+1) + "1"
		f.SetCellValue(sheetName, cell, header)
		// f.SetCellStyle(sheetName, cell, cell, `{"font":{"bold":true}}`) // Set header menjadi bold
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	// Tulis data transaksi
	for rowIndex, transaction := range transactions {
		var status string
		switch transaction.StatusCode {
		case 1005:
			status = "Failed"
		case 1001 | 1003:
			status = "Pending"
		case 1000:
			status = "Success"
		}

		createdAt := transaction.CreatedAt.In(loc).Format("2006-01-02 15:04:05")

		row := rowIndex + 2 // Mulai dari baris kedua setelah header
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), transaction.ID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), createdAt)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), transaction.MtTid)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), transaction.PaymentMethod)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), transaction.Amount)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), transaction.UserId)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), transaction.AppName)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), transaction.Currency)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(row), transaction.ItemName)
		f.SetCellValue(sheetName, "J"+strconv.Itoa(row), transaction.ItemId)
		f.SetCellValue(sheetName, "K"+strconv.Itoa(row), status)
	}

	// Set border untuk header
	// style, err := f.NewStyle(`{"border":[{"type":"thin","color":"#000000","size":1}]}`)
	// if err != nil {
	// 	return err
	// }
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
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &parsedStartDate
		}
	}
	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = &parsedEndDate
		}
	}

	arrClient, err := repository.FindClient(context.Background(), appKey, appID)
	if err != nil {
		fmt.Println("Error fetching client:", err)
	}

	transactions, totalItems, err := repository.GetTransactionsMerchant(context.Background(), limit, offset, status, denom, merchantTransactionId, arrClient.ClientName, userMDN, appName, paymentMethods, startDate, endDate)
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

	// if transaction.StatusCode != 1000 || transaction.StatusCode == 1003 {
	// 	return response.Response(c, fiber.StatusInternalServerError, "Transaction not success")

	// }
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

	if callbackURL == "" {
		log.Printf("No matching ClientApp found for AppID: %s", transaction.AppID)
	}

	statusCode := 1000

	callbackData := repository.CallbackData{
		UserID:                transaction.UserId,
		MerchantTransactionID: transaction.MtTid,
		StatusCode:            statusCode,
		PaymentMethod:         transaction.PaymentMethod,
		Amount:                fmt.Sprintf("%d", transaction.Amount),
		Status:                "success",
		Currency:              transaction.Currency,
		ItemName:              transaction.ItemName,
		ItemID:                transaction.ItemId,
		ReferenceID:           transaction.ReferenceID,
	}

	err = repository.SendCallback(callbackURL, arrClient.ClientSecret, transaction.ID, callbackData)
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
		"message": "Successful MT Transaction",
	})
}
