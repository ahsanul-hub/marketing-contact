package handler

import (
	"app/config"
	"app/repository"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BillRequest struct {
	CompanyCode     string `json:"CompanyCode"`
	CustomerNumber  string `json:"CustomerNumber"`
	RequestID       string `json:"RequestID"`
	ChannelType     string `json:"ChannelType"`
	TransactionDate string `json:"TransactionDate"`
	AdditionalData  string `json:"AdditionalData,omitempty"`
}

type BillResponse struct {
	CompanyCode    string `json:"CompanyCode"`
	CustomerNumber string `json:"CustomerNumber"`
	RequestID      string `json:"RequestID"`
	InquiryStatus  string `json:"InquiryStatus"`
	InquiryReason  *struct {
		Indonesian string `json:"Indonesian,omitempty"`
		English    string `json:"English,omitempty"`
	} `json:"InquiryReason,omitempty"`
	CustomerName   string   `json:"CustomerName,omitempty"`
	CurrencyCode   string   `json:"CurrencyCode,omitempty"`
	TotalAmount    string   `json:"TotalAmount,omitempty"`
	SubCompany     string   `json:"SubCompany,omitempty"`
	DetailBills    []string `json:"DetailBills,omitempty"`
	FreeText       []string `json:"FreeText,omitempty"`
	AdditionalData string   `json:"AdditionalData"`
}

type PaymentRequest struct {
	CompanyCode     string   `json:"CompanyCode"`
	CustomerNumber  string   `json:"CustomerNumber"`
	RequestID       string   `json:"RequestID"`
	ChannelType     string   `json:"ChannelType"`
	CustomerName    string   `json:"CustomerName"`
	CurrencyCode    string   `json:"CurrencyCode"`
	PaidAmount      string   `json:"PaidAmount"`
	TotalAmount     string   `json:"TotalAmount"`
	SubCompany      string   `json:"SubCompany"`
	TransactionDate string   `json:"TransactionDate"`
	Reference       string   `json:"Reference"`
	DetailBills     []string `json:"DetailBills,omitempty"`
	FlagAdvice      string   `json:"FlagAdvice"`
	AdditionalData  string   `json:"AdditionalData,omitempty"`
}

// {
// 	"CompanyCode": "11131",
//    "CustomerNumber": "1592314503",
//    "RequestID": "202503270016541113100898856378",
//    "ChannelType": "6017",
//    "CustomerName":"Ahsanul Waladi",
//    "CurrencyCode":"IDR",
//    "PaidAmount":"10000",
//    "TotalAmount":"10000",
//    "SubCompany":"Redision",
//    "TransactionDate": "22/04/2025 14:10:29",
//    "Reference": "frgtw453564523u",
//    "DetailBills": "",
//    "FlagAdvice": "Y",
//    "AdditionalData": ""
// }

// Struktur response untuk pembayaran
type PaymentResponse struct {
	CompanyCode       string `json:"CompanyCode"`
	CustomerNumber    string `json:"CustomerNumber"`
	RequestID         string `json:"RequestID"`
	PaymentFlagStatus string `json:"PaymentFlagStatus"`
	PaymentFlagReason *struct {
		Indonesian string `json:"Indonesian,omitempty"`
		English    string `json:"English,omitempty"`
	} `json:"PaymentFlagReason,omitempty"`
	CustomerName    string   `json:"CustomerName"`
	CurrencyCode    string   `json:"CurrencyCode,omitempty"`
	PaidAmount      string   `json:"PaidAmount,omitempty"`
	TotalAmount     string   `json:"TotalAmount,omitempty"`
	TransactionDate string   `json:"TransactionDate,omitempty"`
	DetailBills     []string `json:"DetailBills,omitempty"`
	FreeText        []string `json:"FreeText,omitempty"`
	AdditionalData  string   `json:"AdditionalData"`
}

type VaBCAErrorResponse struct {
	ErrorCode    string       `json:"ErrorCode"`
	ErrorMessage ErrorMessage `json:"ErrorMessage"`
}

type ErrorMessage struct {
	Indonesian string `json:"Indonesian"`
	English    string `json:"English"`
}

// func GetRedpayBCAToken() (string, error) {
// 	if cached, found := lib.RedpayTokenCache.Get("redpay_token"); found {
// 		tokenData := cached.(lib.CachedToken)
// 		if time.Now().Before(tokenData.ExpiredAt) {
// 			return tokenData.Token, nil
// 		}
// 	}

// 	tokenResp, err := lib.RequestTokenVaBCARedpay()
// 	if err != nil {
// 		log.Println("error request token lib BCA")
// 		return "", err
// 	}
// 	return tokenResp.AccessToken, nil
// }

var (
	tokenCache     string
	tokenExpiresAt time.Time
	cacheMutex     sync.Mutex
)

func generateRandomToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateOrGetToken() (string, int) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	now := time.Now()

	if now.Before(tokenExpiresAt) && tokenCache != "" {
		expiresIn := int(tokenExpiresAt.Sub(now).Seconds())
		return tokenCache, expiresIn
	}

	tokenCache = generateRandomToken(14)
	tokenExpiresAt = now.Add(3600 * time.Second)
	return tokenCache, 3600
}

func validateBCASignature(c *fiber.Ctx, token, secret, path string) bool {
	// Ambil raw body dan normalisasi
	body := c.Body()
	normalized := bytes.ReplaceAll(body, []byte(" "), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\n"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte(""))
	normalized = bytes.ReplaceAll(normalized, []byte("\t"), []byte(""))

	// SHA256 hash
	hash := sha256.Sum256(normalized)
	bodyHash := hex.EncodeToString(hash[:])

	// Ambil header
	timestamp := c.Get("X-BCA-Timestamp")
	signatureFromHeader := c.Get("X-BCA-Signature")

	// Bangun string to sign
	stringToSign := fmt.Sprintf("POST:/bca/%s:%s:%s:%s", path, token, bodyHash, timestamp)

	// Generate HMAC SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	log.Println("BCA Request Body:\n", string(body))
	log.Println("expected signature", expectedSignature)

	// Bandingkan
	return hmac.Equal([]byte(signatureFromHeader), []byte(expectedSignature))
}

func InquiryBca(c *fiber.Ctx) error {
	var resError VaBCAErrorResponse

	authorization := c.Get("Authorization")
	x_bca_key := c.Get("X-BCA-Key")
	// x_bca_signature := c.Get("X-BCA-Signature")
	// x_bca_timestamp := c.Get("X-BCA-Timestamp")
	secret := "jokwFlBC80WNVCJ"

	token, _ := GenerateOrGetToken()

	expectedAuthorization := fmt.Sprintf("Bearer %s", token)
	expectedXBCAKEy := "XrPd1pztIr"

	// log.Println("authorization", authorization)
	// log.Println("x_bca_key", x_bca_key)
	// log.Println("x_bca_signature", x_bca_signature)
	// log.Println("x_bca_timestamp", x_bca_timestamp)

	if authorization != expectedAuthorization || x_bca_key != expectedXBCAKEy {
		resError = VaBCAErrorResponse{
			ErrorCode: "ERROR-INVALID-AUTHORIZATION",
			ErrorMessage: ErrorMessage{
				Indonesian: "client_id/client_secret tidak valid",
				English:    "Invalid client_id/client_secret",
			},
		}
		return c.Status(fiber.StatusUnauthorized).JSON(resError)
	}

	if !validateBCASignature(c, token, secret, "inquiry") {
		resError = VaBCAErrorResponse{
			ErrorCode: "INVALID_SIGNATURE",
			ErrorMessage: ErrorMessage{
				Indonesian: "Signature tidak valid",
				English:    "Invalid signature",
			},
		}
		return c.Status(fiber.StatusUnauthorized).JSON(resError)
	}

	// log.Println("token pass")
	var request BillRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var response BillResponse

	vaNumber := fmt.Sprintf("11131%s", request.CustomerNumber)

	env := config.Config("ENV", "")

	if env == "development" {
		activeVAs := map[string]string{
			"111316829726801": "Budi Santoso",
			"111319869285781": "Siti Aminah",
			"111311682920749": "Andi Wijaya",
			"111312959571097": "Dewi Lestari",
			"111313846905385": "Rudi Hartono",
			"111314970286953": "Fitriani",
			"111316810136869": "Joko Anwar",
			"111319235817402": "Linda Marbun",
			"111314189406863": "Taufik Hidayat",
			"111313258468127": "Nur Aini",
		}

		expiredVAs := map[string]string{
			"111311326580369": "Expired Darwin",
			"111314672968104": "Expired Louise",
		}

		// Periksa VA dummy aktif
		if name, ok := activeVAs[vaNumber]; ok {
			response := BillResponse{
				CompanyCode:    request.CompanyCode,
				CustomerNumber: request.CustomerNumber,
				RequestID:      request.RequestID,
				InquiryStatus:  "00",
				InquiryReason: &struct {
					Indonesian string `json:"Indonesian,omitempty"`
					English    string `json:"English,omitempty"`
				}{
					Indonesian: "Sukses",
					English:    "Success",
				},
				CustomerName:   name,
				CurrencyCode:   "IDR",
				TotalAmount:    "10000.00",
				SubCompany:     "00000",
				DetailBills:    []string{},
				FreeText:       []string{},
				AdditionalData: "",
			}
			return c.Status(fiber.StatusOK).JSON(response)
		}

		if _, ok := expiredVAs[vaNumber]; ok {
			response := BillResponse{
				CompanyCode:    request.CompanyCode,
				CustomerNumber: request.CustomerNumber,
				RequestID:      request.RequestID,
				InquiryStatus:  "01",
				InquiryReason: &struct {
					Indonesian string `json:"Indonesian,omitempty"`
					English    string `json:"English,omitempty"`
				}{
					Indonesian: "Nomor VA tidak valid atau expired",
					English:    "Invalid VA number or expired",
				},
				CurrencyCode: "IDR",
				TotalAmount:  "10000.00",
				SubCompany:   "00000",
			}
			return c.Status(fiber.StatusOK).JSON(response)
		}
	}

	transaction, err := repository.GetTransactionVa(context.Background(), vaNumber)
	if err != nil {
		response = BillResponse{
			CompanyCode:    request.CompanyCode,
			CustomerNumber: request.CustomerNumber,
			RequestID:      request.RequestID,
			InquiryStatus:  "01",
			InquiryReason: &struct {
				Indonesian string `json:"Indonesian,omitempty"`
				English    string `json:"English,omitempty"`
			}{
				Indonesian: "Nomor VA tidak valid atau expired",
				English:    "Invalid VA number or expired",
			},
			CurrencyCode: "IDR",
			TotalAmount:  "",
			SubCompany:   "00000",
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	totalAmount := fmt.Sprintf("%d.00", transaction.Amount)
	response = BillResponse{
		CompanyCode:    request.CompanyCode,
		CustomerNumber: request.CustomerNumber,
		RequestID:      request.RequestID,
		InquiryStatus:  "00",
		InquiryReason: &struct {
			Indonesian string `json:"Indonesian,omitempty"`
			English    string `json:"English,omitempty"`
		}{
			Indonesian: "Sukses",
			English:    "Success",
		},
		CustomerName:   transaction.CustomerName,
		CurrencyCode:   "IDR",
		TotalAmount:    totalAmount,
		SubCompany:     "00000",
		DetailBills:    []string{},
		FreeText:       []string{},
		AdditionalData: "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func TokenBca(c *fiber.Ctx) error {

	authorization := c.Get("Authorization")
	expectedAuthorization := "Basic UjNkMXMxMG46YXRkc1Vxcml3MTQxQVQzTDlQNFo="

	var resError VaBCAErrorResponse

	if authorization != expectedAuthorization {
		resError = VaBCAErrorResponse{
			ErrorCode: "ERROR-INVALID-AUTHORIZATION",
			ErrorMessage: ErrorMessage{
				Indonesian: "client_id/client_secret tidak valid",
				English:    "Invalid client_id/client_secret",
			},
		}
		return c.Status(fiber.StatusOK).JSON(resError)
	}

	// tokenRedpay, err := lib.RequestTokenVaBCARedpay()
	// if err != nil {
	// 	log.Println("err", err)
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error request token"})
	// }

	token, expiresIn := GenerateOrGetToken()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": token,
		"expires_in":   expiresIn,
	})

}

func PaymentBca(c *fiber.Ctx) error {

	// log.Println("Headers:")
	// c.Request().Header.VisitAll(func(key, value []byte) {
	// 	fmt.Printf("%s: %s\n", key, value)
	// })

	// Log Body
	body := c.Body()
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		log.Println("Body (raw):", string(body))
	} else {
		log.Println("va bca payment Body  (formatted):")
		fmt.Println(prettyJSON.String())
	}
	var request PaymentRequest
	if err := c.BodyParser(&request); err != nil {
		log.Println("Invalid request payment va_bca")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var response PaymentResponse

	authorization := c.Get("Authorization")
	x_bca_key := c.Get("X-BCA-Key")
	secret := "jokwFlBC80WNVCJ"

	var resError VaBCAErrorResponse

	token, _ := GenerateOrGetToken()

	expectedAuthorization := fmt.Sprintf("Bearer %s", token)
	expectedXBCAKEy := "XrPd1pztIr"

	log.Println("expectedAuthorization", expectedAuthorization)

	if authorization != expectedAuthorization || x_bca_key != expectedXBCAKEy {
		resError = VaBCAErrorResponse{
			ErrorCode: "ERROR-INVALID-AUTHORIZATION",
			ErrorMessage: ErrorMessage{
				Indonesian: "client_id/client_secret tidak valid",
				English:    "Invalid client_id/client_secret",
			},
		}
		return c.Status(fiber.StatusUnauthorized).JSON(resError)
	}

	if !validateBCASignature(c, token, secret, "payment") {
		resError = VaBCAErrorResponse{
			ErrorCode: "INVALID_SIGNATURE",
			ErrorMessage: ErrorMessage{
				Indonesian: "Signature tidak valid",
				English:    "Invalid signature",
			},
		}
		return c.Status(fiber.StatusUnauthorized).JSON(resError)
	}

	vaNumber := fmt.Sprintf("11131%s", request.CustomerNumber)

	paidFloat, err := strconv.ParseFloat(request.PaidAmount, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid PaidAmount"})
	}
	paidAmount := uint(paidFloat)

	// totalFloat, err := strconv.ParseFloat(request.PaidAmount, 64)
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Total"})
	// }
	// totalAmount := uint(totalFloat)

	env := config.Config("ENV", "")

	if env == "development" {
		activeVAs := map[string]string{
			"111316829726801": "Budi Santoso",
			"111319869285781": "Siti Aminah",
			"111311682920749": "Andi Wijaya",
			"111312959571097": "Dewi Lestari",
			"111313846905385": "Rudi Hartono",
			"111314970286953": "Fitriani",
			"111316810136869": "Joko Anwar",
			"111319235817402": "Linda Marbun",
			"111314189406863": "Taufik Hidayat",
			"111313258468127": "Nur Aini",
		}

		expiredVAs := map[string]string{
			"111311326580369": "Expired Darwin",
			"111314672968104": "Expired Louise",
		}

		// Periksa VA dummy aktif
		if _, ok := activeVAs[vaNumber]; ok {
			response = PaymentResponse{
				CompanyCode:       request.CompanyCode,
				CustomerNumber:    request.CustomerNumber,
				RequestID:         request.RequestID,
				PaymentFlagStatus: "00",
				PaymentFlagReason: &struct {
					Indonesian string `json:"Indonesian,omitempty"`
					English    string `json:"English,omitempty"`
				}{
					Indonesian: "Sukses",
					English:    "Success",
				},
				CustomerName:    request.CustomerName,
				CurrencyCode:    "IDR",
				PaidAmount:      request.PaidAmount,
				TotalAmount:     "10000.00",
				TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
				DetailBills:     []string{},
				FreeText:        []string{},
				AdditionalData:  "",
			}
			return c.Status(fiber.StatusOK).JSON(response)
		}

		if _, ok := expiredVAs[vaNumber]; ok {
			response = PaymentResponse{
				CompanyCode:       request.CompanyCode,
				CustomerNumber:    request.CustomerNumber,
				RequestID:         request.RequestID,
				PaymentFlagStatus: "01",
				PaymentFlagReason: &struct {
					Indonesian string `json:"Indonesian,omitempty"`
					English    string `json:"English,omitempty"`
				}{
					Indonesian: "Payment tidak valid",
					English:    "Invalid payment",
				},
				CurrencyCode:    "IDR",
				PaidAmount:      request.PaidAmount,
				TotalAmount:     "10000.00",
				TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
				DetailBills:     []string{},
				FreeText:        []string{},
				AdditionalData:  "",
			}
			return c.Status(fiber.StatusOK).JSON(response)
		}
	}

	transaction, err := repository.GetTransactionVa(context.Background(), vaNumber)
	if err != nil {
		response = PaymentResponse{
			CompanyCode:       request.CompanyCode,
			CustomerNumber:    request.CustomerNumber,
			RequestID:         request.RequestID,
			PaymentFlagStatus: "01",
			PaymentFlagReason: &struct {
				Indonesian string `json:"Indonesian,omitempty"`
				English    string `json:"English,omitempty"`
			}{
				Indonesian: "Nomor VA tidak valid",
				English:    "Invalid VA number",
			},
			CurrencyCode:    "IDR",
			PaidAmount:      request.PaidAmount,
			TotalAmount:     fmt.Sprintf("%d.00", transaction.Amount),
			TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
			DetailBills:     []string{},
			FreeText:        []string{},
			AdditionalData:  "",
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}

	timeLimit := time.Now().Add(-70 * time.Minute)
	if transaction.CreatedAt.Before(timeLimit) {
		response = PaymentResponse{
			CompanyCode:       request.CompanyCode,
			CustomerNumber:    request.CustomerNumber,
			RequestID:         request.RequestID,
			PaymentFlagStatus: "01",
			PaymentFlagReason: &struct {
				Indonesian string `json:"Indonesian,omitempty"`
				English    string `json:"English,omitempty"`
			}{
				Indonesian: "Payment tidak valid",
				English:    "Invalid payment",
			},
			CurrencyCode:    "IDR",
			PaidAmount:      request.PaidAmount,
			TotalAmount:     fmt.Sprintf("%d.00", transaction.Amount),
			TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
			DetailBills:     []string{},
			FreeText:        []string{},
			AdditionalData:  "",
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}

	if paidAmount != transaction.Amount {
		response = PaymentResponse{
			CompanyCode:       request.CompanyCode,
			CustomerNumber:    request.CustomerNumber,
			RequestID:         request.RequestID,
			PaymentFlagStatus: "01",
			PaymentFlagReason: &struct {
				Indonesian string `json:"Indonesian,omitempty"`
				English    string `json:"English,omitempty"`
			}{
				Indonesian: "Payment tidak valid",
				English:    "Invalid payment",
			},
			CurrencyCode:    "IDR",
			PaidAmount:      request.PaidAmount,
			TotalAmount:     fmt.Sprintf("%d.00", transaction.Amount),
			TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
			DetailBills:     []string{},
			FreeText:        []string{},
			AdditionalData:  "",
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	now := time.Now()

	receiveCallbackDate := &now

	// log.Println("token pass")
	err = repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, &request.RequestID, nil, "", receiveCallbackDate)
	if err != nil {
		log.Println("failed update status success va_bca")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update transaction status"})
	}

	response = PaymentResponse{
		CompanyCode:       request.CompanyCode,
		CustomerNumber:    request.CustomerNumber,
		RequestID:         request.RequestID,
		PaymentFlagStatus: "00",
		PaymentFlagReason: &struct {
			Indonesian string `json:"Indonesian,omitempty"`
			English    string `json:"English,omitempty"`
		}{
			Indonesian: "Sukses",
			English:    "Success",
		},
		CustomerName:    request.CustomerName,
		CurrencyCode:    "IDR",
		PaidAmount:      request.PaidAmount,
		TotalAmount:     fmt.Sprintf("%d.00", transaction.Amount),
		TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
		DetailBills:     []string{},
		FreeText:        []string{},
		AdditionalData:  "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
