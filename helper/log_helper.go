package helper

import (
	"app/config"
	"encoding/json"
	"fmt"
	"time"
)

// PaymentHelpers provides logging helpers for payment libraries
type PaymentHelpers struct {
	PaymentMethod string
}

// NewPaymentHelpers creates a new instance for a specific payment method
func NewPaymentHelpers(paymentMethod string) *PaymentHelpers {
	return &PaymentHelpers{
		PaymentMethod: paymentMethod,
	}
}

// LogTransactionStart - only log if needed for debugging critical issues
func (ph *PaymentHelpers) LogTransactionStart(transactionID, userID string, amount uint, data map[string]interface{}) {
	// Skip routine transaction starts to reduce log volume
}

// LogTransactionSuccess - skip success logs to reduce volume
func (ph *PaymentHelpers) LogTransactionSuccess(transactionID, userID string, amount uint, data map[string]interface{}) {
	// Skip success logs to reduce volume
}

// LogTransactionError logs transaction error - this is important
func (ph *PaymentHelpers) LogTransactionError(transactionID, userID string, amount uint, errorMsg string, data map[string]interface{}) {
	entry := config.LogEntry{
		TransactionID: transactionID,
		UserID:        userID,
		Amount:        amount,
		Status:        "error",
		Error:         errorMsg,
		Data:          data,
	}
	config.LogError(ph.PaymentMethod, "Transaction failed", entry)
}

// LogAPICall logs external API calls
func (ph *PaymentHelpers) LogAPICall(endpoint, method string, duration time.Duration, statusCode int, requestData, responseData map[string]interface{}) {
	data := map[string]interface{}{
		"endpoint":    endpoint,
		"method":      method,
		"status_code": statusCode,
	}

	if requestData != nil {
		data["request"] = requestData
	}

	if responseData != nil {
		data["response"] = responseData
	}

	config.LogPaymentAPI(ph.PaymentMethod, endpoint, method, duration, statusCode, data)
}

// LogCallback logs callback received from payment provider - both success and failure
func (ph *PaymentHelpers) LogCallback(transactionID string, success bool, callbackData map[string]interface{}) {
	config.LogPaymentCallback(ph.PaymentMethod, transactionID, success, callbackData)
}

// LogWithData logs with custom data - only WARN and ERROR
func (ph *PaymentHelpers) LogWithData(level, message string, data map[string]interface{}) {
	if level == "WARN" || level == "ERROR" {
		entry := config.LogEntry{
			Data: data,
		}
		config.LogManager.LogPayment(ph.PaymentMethod, level, message, entry)
	}
}

// LogJSON logs JSON data - only for errors
func (ph *PaymentHelpers) LogJSON(level, message string, jsonData interface{}) {
	if level != "ERROR" {
		return
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		ph.LogWithData("ERROR", fmt.Sprintf("Failed to marshal JSON for %s: %v", message, err), nil)
		return
	}

	data := map[string]interface{}{
		"json_data": string(jsonBytes),
	}

	ph.LogWithData(level, message, data)
}

// LogValidation logs validation errors - only failures
func (ph *PaymentHelpers) LogValidation(transactionID string, isValid bool, validationErrors []string, data map[string]interface{}) {
	if isValid {
		return // Skip valid validation logs
	}

	status := "invalid"
	level := "WARN"
	message := "Validation failed"

	logData := map[string]interface{}{
		"validation_status": status,
	}

	if len(validationErrors) > 0 {
		logData["validation_errors"] = validationErrors
	}

	for k, v := range data {
		logData[k] = v
	}

	entry := config.LogEntry{
		TransactionID: transactionID,
		Status:        status,
		Data:          logData,
	}

	config.LogManager.LogPayment(ph.PaymentMethod, level, message, entry)
}

// LogRetry logs retry attempts for failed operations
func (ph *PaymentHelpers) LogRetry(transactionID string, attempt int, maxAttempts int, lastError string, data map[string]interface{}) {
	logData := map[string]interface{}{
		"attempt":      attempt,
		"max_attempts": maxAttempts,
		"last_error":   lastError,
	}

	for k, v := range data {
		logData[k] = v
	}

	entry := config.LogEntry{
		TransactionID: transactionID,
		Error:         lastError,
		Data:          logData,
	}

	level := "WARN"
	message := fmt.Sprintf("Retry attempt %d/%d", attempt, maxAttempts)

	if attempt >= maxAttempts {
		level = "ERROR"
		message = "Max retry attempts reached"
	}

	config.LogManager.LogPayment(ph.PaymentMethod, level, message, entry)
}

// LogConfiguration - skip routine config logs
func (ph *PaymentHelpers) LogConfiguration(configName string, configData map[string]interface{}) {
	// Skip routine configuration logs to reduce volume
}

// Helper functions for specific payment methods
var (
	TelkomselLogger    = NewPaymentHelpers(config.PAYMENT_TELKOMSEL)
	XLLogger           = NewPaymentHelpers(config.PAYMENT_XL)
	IndosatLogger      = NewPaymentHelpers(config.PAYMENT_INDOSAT)
	TriLogger          = NewPaymentHelpers(config.PAYMENT_TRI)
	SmartfrenLogger    = NewPaymentHelpers(config.PAYMENT_SMARTFREN)
	DanaLogger         = NewPaymentHelpers(config.PAYMENT_DANA)
	OvoLogger          = NewPaymentHelpers(config.PAYMENT_OVO)
	GopayLogger        = NewPaymentHelpers(config.PAYMENT_GOPAY)
	ShopeepayLogger    = NewPaymentHelpers(config.PAYMENT_SHOPEEPAY)
	QrisLogger         = NewPaymentHelpers(config.PAYMENT_QRIS)
	VaBcaLogger        = NewPaymentHelpers(config.PAYMENT_VA_BCA)
	VaBniLogger        = NewPaymentHelpers(config.PAYMENT_VA_BNI)
	VaBriLogger        = NewPaymentHelpers(config.PAYMENT_VA_BRI)
	VaMandiriLogger    = NewPaymentHelpers(config.PAYMENT_VA_MANDIRI)
	VaPermataLogger    = NewPaymentHelpers(config.PAYMENT_VA_PERMATA)
	VaSinarmasLogger   = NewPaymentHelpers(config.PAYMENT_VA_SINARMAS)
	MidtransLogger     = NewPaymentHelpers(config.PAYMENT_MIDTRANS)
	HarsyaLogger       = NewPaymentHelpers(config.PAYMENT_HARSYA)
	FaspayLogger       = NewPaymentHelpers(config.PAYMENT_FASPAY)
	TriyakomLogger     = NewPaymentHelpers(config.PAYMENT_TRIYAKOM)
	NotificationLogger = NewPaymentHelpers(config.PAYMENT_NOTIFICATION)
)
