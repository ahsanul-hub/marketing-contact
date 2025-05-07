package handler

import (
	"app/lib"
	"app/repository"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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
	CompanyCode     string `json:"CompanyCode"`
	CustomerNumber  string `json:"CustomerNumber"`
	RequestID       string `json:"RequestID"`
	ChannelType     string `json:"ChannelType"`
	CustomerName    string `json:"CustomerName"`
	CurrencyCode    string `json:"CurrencyCode"`
	PaidAmount      uint   `json:"PaidAmount"`
	TotalAmount     uint   `json:"TotalAmount"`
	SubCompany      string `json:"SubCompany"`
	TransactionDate string `json:"TransactionDate"`
	Reference       string `json:"Reference"`
	DetailBills     string `json:"DetailBills,omitempty"`
	FlagAdvice      string `json:"FlagAdvice"`
	AdditionalData  string `json:"AdditionalData,omitempty"`
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
	PaidAmount      uint     `json:"PaidAmount,omitempty"`
	TotalAmount     uint     `json:"TotalAmount,omitempty"`
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

func GetRedpayBCAToken() (string, error) {
	if cached, found := lib.RedpayTokenCache.Get("redpay_token"); found {
		tokenData := cached.(lib.CachedToken)
		if time.Now().Before(tokenData.ExpiredAt) {
			return tokenData.Token, nil
		}
	}

	tokenResp, err := lib.RequestTokenVaBCARedpay()
	if err != nil {
		log.Println("error request token lib BCA")
		return "", err
	}
	return tokenResp.AccessToken, nil
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

	token, err := GetRedpayBCAToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error get token redpay bca"})
	}
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

	// data, found := VaTransactionCache.Get(vaNumber)
	// if !found {
	// 	return c.Status(fiber.StatusNotFound).SendString("Transaction not found or expired")
	// }
	// log.Println("dataVa", data)

	// Pecah qrisUrl dan acquirer
	// dataStr := data.(string)
	// parts := strings.Split(dataStr, "|")
	// if len(parts) != 3 {
	// 	return c.Status(fiber.StatusInternalServerError).SendString("Invalid data format")
	// }
	// transactionID := parts[0]

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
			TotalAmount:  fmt.Sprintf("%d.00", transaction.Amount),
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

	tokenRedpay, err := lib.RequestTokenVaBCARedpay()
	if err != nil {
		log.Println(tokenRedpay)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error request token"})
	}

	return c.Status(fiber.StatusOK).JSON(tokenRedpay)
}

func PaymentBca(c *fiber.Ctx) error {
	log.Println("===== [BCA Payment] Incoming Request =====")

	// Log Header
	log.Println("Headers:")
	c.Request().Header.VisitAll(func(key, value []byte) {
		fmt.Printf("%s: %s\n", key, value)
	})

	// Log Body
	body := c.Body()
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		log.Println("Body (raw):", string(body))
	} else {
		log.Println("Body (formatted):")
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

	token, err := GetRedpayBCAToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error get token redpay bca"})
	}
	expectedAuthorization := fmt.Sprintf("Bearer %s", token)
	expectedXBCAKEy := "XrPd1pztIr"

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
			TotalAmount:     transaction.Amount,
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
			TotalAmount:     transaction.Amount,
			TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
			DetailBills:     []string{},
			FreeText:        []string{},
			AdditionalData:  "",
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
	// log.Println("request.PaidAmount", request.PaidAmount)
	// log.Println("amount", transaction.Amount)

	if request.PaidAmount != transaction.Amount {
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
			TotalAmount:     transaction.Amount,
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
	err = repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, nil, nil, "", receiveCallbackDate)
	if err != nil {
		log.Println("failed update status success va_bca")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update transaction status"})
	}

	log.Println("success update status")

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
		TotalAmount:     transaction.Amount,
		TransactionDate: time.Now().Format("02/01/2006 15:04:05"),
		DetailBills:     []string{},
		FreeText:        []string{},
		AdditionalData:  "",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
