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
	"log"
	"math"
	"strconv"
	"time"

	"fmt"

	"github.com/gofiber/fiber/v2"
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
	arrClient, err := repository.FindClient(spanCtx, c.Get("appkey"), c.Get("appid"))

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}
	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)
	createdTransId, err := repository.CreateTransaction(context.Background(), &transaction, arrClient)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	switch transaction.PaymentMethod {
	case "xl_airtime":
		chargingResponse, err := lib.RequestCharging(transaction.UserMDN, transaction.MtTid, transaction.ItemName, createdTransId, transaction.Amount)
		if err != nil {

			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Charging successful",
			"data":    chargingResponse,
		})
	case "smartfren":

	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Transaction created successfully")
}

func CreateTransactionV1(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "CreateTransactionV1", "handler")
	defer span.End()

	token := c.Get("token")

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
	arrClient, err := repository.FindClient(spanCtx, c.Get("appkey"), c.Get("appid"))

	transaction.UserMDN = helper.BeautifyIDNumber(transaction.UserMDN, true)

	if err != nil {
		return response.Response(c, fiber.StatusBadRequest, "E0001")
	}

	createdTransId, err := repository.CreateTransaction(spanCtx, &transaction, arrClient)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	switch transaction.PaymentMethod {
	case "xl_airtime":
		chargingResponse, err := lib.RequestCharging(transaction.UserMDN, transaction.MtTid, transaction.ItemName, createdTransId, transaction.Amount)
		if err != nil {

			log.Println("Charging request failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Charging request failed",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Charging successful",
			"token":   token,
			"data":    chargingResponse,
		})
	case "smartfren":

	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Transaction created successfully")
}

func GetTransactions(c *fiber.Ctx) error {
	span, spanCtx := apm.StartSpan(c.Context(), "GetTransactions", "handler")
	defer span.End()

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	appID := c.Query("app_id")
	userMDN := helper.BeautifyIDNumber(c.Query("user_mdn"), true)
	log.Println(userMDN)
	paymentMethod := c.Query("payment_method")
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

	transactions, err := repository.GetAllTransactions(spanCtx, limit, offset, appID, userMDN, paymentMethod, startDate, endDate)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transactions,
	})
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

	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	userMDN := c.Query("user_mdn")
	paymentMethod := c.Query("payment_method")
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

	transactions, err := repository.GetTransactionsMerchant(context.Background(), limit, offset, appKey, appID, userMDN, paymentMethod, startDate, endDate)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"success": true, "data": transactions})
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
