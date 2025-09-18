package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PaymentLogEntry represents a log entry for payment-specific operations
type LogEntry struct {
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	Timestamp     string                 `json:"timestamp"`
	PaymentType   string                 `json:"payment_type"`
	TransactionID string                 `json:"transaction_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Amount        uint                   `json:"amount,omitempty"`
	Status        string                 `json:"status,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Duration      float64                `json:"duration_ms,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// PaymentLogger manages logging for specific payment methods
type PaymentLogger struct {
	paymentType string
	logger      *log.Logger
	logFile     *os.File
	logChan     chan LogEntry
	stopChan    chan bool
	wg          sync.WaitGroup
}

// PaymentLoggerManager manages all payment loggers
type LoggerManager struct {
	loggers map[string]*PaymentLogger
	mu      sync.RWMutex
}

var (
	LogManager *LoggerManager
)

// Payment method constants
const (
	PAYMENT_TELKOMSEL   = "telkomsel"
	PAYMENT_XL          = "xl"
	PAYMENT_INDOSAT     = "indosat"
	PAYMENT_TRI         = "tri"
	PAYMENT_SMARTFREN   = "smartfren"
	PAYMENT_DANA        = "dana"
	PAYMENT_OVO         = "ovo"
	PAYMENT_GOPAY       = "gopay"
	PAYMENT_SHOPEEPAY   = "shopeepay"
	PAYMENT_QRIS        = "qris"
	PAYMENT_VA_BCA      = "va_bca"
	PAYMENT_VA_BNI      = "va_bni"
	PAYMENT_VA_BRI      = "va_bri"
	PAYMENT_VA_MANDIRI  = "va_mandiri"
	PAYMENT_VA_PERMATA  = "va_permata"
	PAYMENT_VA_SINARMAS = "va_sinarmas"
	PAYMENT_MIDTRANS    = "midtrans"
	PAYMENT_HARSYA      = "harsya"
	PAYMENT_FASPAY      = "faspay"
	PAYMENT_TRIYAKOM    = "triyakom"
	PAYMENT_CALLBACK    = "callback"
)

// InitPaymentLoggers initializes all payment method loggers
func InitPaymentLoggers() error {
	LogManager = &LoggerManager{
		loggers: make(map[string]*PaymentLogger),
	}

	// List of all payment methods to create loggers for
	paymentMethods := []string{
		PAYMENT_TELKOMSEL, PAYMENT_XL, PAYMENT_INDOSAT, PAYMENT_TRI, PAYMENT_SMARTFREN,
		PAYMENT_DANA, PAYMENT_OVO, PAYMENT_GOPAY, PAYMENT_SHOPEEPAY, PAYMENT_QRIS,
		PAYMENT_VA_BCA, PAYMENT_VA_BNI, PAYMENT_VA_BRI, PAYMENT_VA_MANDIRI,
		PAYMENT_VA_PERMATA, PAYMENT_VA_SINARMAS, PAYMENT_MIDTRANS, PAYMENT_HARSYA,
		PAYMENT_FASPAY, PAYMENT_TRIYAKOM, PAYMENT_CALLBACK, "admin", "auth",
	}

	for _, method := range paymentMethods {
		if err := LogManager.CreateLogger(method); err != nil {
			return fmt.Errorf("failed to create logger for %s: %w", method, err)
		}
	}

	return nil
}

// CreateLogger creates a new payment logger with goroutine worker
func (plm *LoggerManager) CreateLogger(paymentType string) error {
	plm.mu.Lock()
	defer plm.mu.Unlock()

	// Ensure logs directory is under project root, regardless of CWD
	baseDir := LogsBaseDir()
	logDir := filepath.Join(baseDir, "payments")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	currentTime := time.Now()
	year, month, _ := currentTime.Date()
	_, week := currentTime.ISOWeek()

	logFileName := fmt.Sprintf("dcb-%d-%02d-week%d-%s.log", year, month, week, paymentType)
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logger := log.New(logFile, "", 0)

	paymentLogger := &PaymentLogger{
		paymentType: paymentType,
		logger:      logger,
		logFile:     logFile,
		logChan:     make(chan LogEntry, 1000), // Buffer untuk 1000 log entries
		stopChan:    make(chan bool),
	}

	// Start goroutine worker untuk logger ini
	paymentLogger.wg.Add(1)
	go paymentLogger.worker()

	plm.loggers[paymentType] = paymentLogger
	return nil
}

// worker is the goroutine that processes log entries
func (pl *PaymentLogger) worker() {
	defer pl.wg.Done()

	for {
		select {
		case entry := <-pl.logChan:
			pl.writeLog(entry)
		case <-pl.stopChan:
			// Process remaining entries before stopping
			for len(pl.logChan) > 0 {
				entry := <-pl.logChan
				pl.writeLog(entry)
			}
			return
		}
	}
}

// writeLog writes the log entry to file
func (pl *PaymentLogger) writeLog(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshaling log entry for %s: %v", pl.paymentType, err)
		return
	}

	pl.logger.Println(string(jsonData))
	pl.logFile.Sync() // Force write to disk
}

// LogPayment logs a payment-related event asynchronously
func (plm *LoggerManager) LogPayment(paymentType, level, message string, entry LogEntry) {
	// Option 2: Balanced - log ERROR, WARN, and selective INFO
	if level != "ERROR" && level != "WARN" && level != "INFO" {
		return
	}

	plm.mu.RLock()
	logger, exists := plm.loggers[paymentType]
	plm.mu.RUnlock()

	if !exists {
		// Fallback ke logger utama jika logger spesifik payment tidak ditemukan
		log.Printf("WARN: Payment logger untuk %s tidak ditemukan", paymentType)
		return
	}
	entry.Level = level
	entry.Message = message
	entry.PaymentType = paymentType
	entry.Timestamp = time.Now().Format(time.RFC3339)

	// Send to goroutine channel (non-blocking)
	select {
	case logger.logChan <- entry:
		// Successfully queued
	default:
		// Channel is full, log to main logger as fallback
		LogWithLevel("ERROR", "Payment logger channel full for %s", paymentType)
	}
}

// Convenience functions for different log levels - Balanced approach
func LogInfo(paymentType, message string, entry LogEntry) {
	// Only log success transactions and important events
	if entry.Status == "success" || entry.Status == "paid" || entry.Status == "completed" {
		LogManager.LogPayment(paymentType, "INFO", message, entry)
	}
}

func LogError(paymentType, message string, entry LogEntry) {
	LogManager.LogPayment(paymentType, "ERROR", message, entry)
}

func LogPaymentWarn(paymentType, message string, entry LogEntry) {
	LogManager.LogPayment(paymentType, "WARN", message, entry)
}

func LogDebug(paymentType, message string, entry LogEntry) {
	// Skip DEBUG logs to reduce volume
}

// Convenience functions for specific payment operations - Balanced logging
func LogPaymentTransaction(paymentType, transactionID, userID string, amount uint, status string, data map[string]interface{}) {
	entry := LogEntry{
		TransactionID: transactionID,
		UserID:        userID,
		Amount:        amount,
		Status:        status,
		Data:          data,
	}

	// Log both success and failure for complete transaction tracking
	if status == "error" || status == "failed" {
		LogError(paymentType, "Transaction failed", entry)
	} else if status == "success" || status == "paid" || status == "completed" {
		LogInfo(paymentType, "Transaction completed", entry)
	}
}

func LogPaymentCallback(paymentType, transactionID string, success bool, data map[string]interface{}) {
	entry := LogEntry{
		TransactionID: transactionID,
		Data:          data,
	}

	// Log both success and failed callbacks
	if !success {
		entry.Status = "failed"
		LogManager.LogPayment(paymentType, "ERROR", "Callback failed", entry)
	} else {
		entry.Status = "success"
		LogManager.LogPayment(paymentType, "INFO", "Callback success", entry)
	}
}

func LogPaymentAPI(paymentType, endpoint, method string, duration time.Duration, statusCode int, data map[string]interface{}) {
	entry := LogEntry{
		Duration: float64(duration.Nanoseconds()) / 1e6, // Convert to milliseconds
		Data: map[string]interface{}{
			"endpoint":    endpoint,
			"method":      method,
			"status_code": statusCode,
		},
	}

	for k, v := range data {
		entry.Data[k] = v
	}

	// message := fmt.Sprintf("API call Duration: %.2fms)", entry.Duration)

	// Balanced logging approach
	if statusCode >= 400 {
		LogManager.LogPayment(paymentType, "ERROR", "", entry)
	} else if duration > 2*time.Second {
		LogManager.LogPayment(paymentType, "WARN", "", entry)
	} else if statusCode == 200 || statusCode == 201 {
		// Log semua transaksi sukses (200/201) ke file payment method
		LogManager.LogPayment(paymentType, "INFO", "", entry)
	}
	// Skip other normal responses to reduce volume
}

// ShutdownPaymentLoggers gracefully shuts down all payment loggers
func ShutdownPaymentLoggers() {
	if LogManager == nil {
		return
	}

	LogManager.mu.Lock()
	defer LogManager.mu.Unlock()

	// Silent shutdown - only log if there are errors
	for _, logger := range LogManager.loggers {
		close(logger.stopChan)
		logger.wg.Wait()
		logger.logFile.Close()
	}

	// Single log message for successful shutdown
	LogWithLevel("INFO", "Payment loggers shutdown completed")
}

// GetPaymentMethodFromEndpoint extracts payment method from endpoint or payment method string
func GetPaymentMethodFromEndpoint(endpoint, paymentMethod string) string {
	// Map common endpoint patterns to payment methods
	endpointMap := map[string]string{
		"/telkomsel": PAYMENT_TELKOMSEL,
		"/xl":        PAYMENT_XL,
		"/indosat":   PAYMENT_INDOSAT,
		"/tri":       PAYMENT_TRI,
		"/smartfren": PAYMENT_SMARTFREN,
		"/dana":      PAYMENT_DANA,
		"/ovo":       PAYMENT_OVO,
		"/gopay":     PAYMENT_GOPAY,
		"/shopeepay": PAYMENT_SHOPEEPAY,
		"/qris":      PAYMENT_QRIS,
		"/bca":       PAYMENT_VA_BCA,
		"/bni":       PAYMENT_VA_BNI,
		"/bri":       PAYMENT_VA_BRI,
		"/mandiri":   PAYMENT_VA_MANDIRI,
		"/permata":   PAYMENT_VA_PERMATA,
		"/sinarmas":  PAYMENT_VA_SINARMAS,
		"/midtrans":  PAYMENT_MIDTRANS,
		"/harsya":    PAYMENT_HARSYA,
		"/faspay":    PAYMENT_FASPAY,
		"/triyakom":  PAYMENT_TRIYAKOM,
		"/callback":  PAYMENT_CALLBACK, // Generic callback
	}

	// First check if paymentMethod directly matches our constants
	switch paymentMethod {
	case "telkomsel", "telkomsel_airtime":
		return PAYMENT_TELKOMSEL
	case "xl", "xl_airtime":
		return PAYMENT_XL
	case "indosat", "indosat_airtime":
		return PAYMENT_INDOSAT
	case "tri", "tri_airtime":
		return PAYMENT_TRI
	case "smartfren", "smartfren_airtime":
		return PAYMENT_SMARTFREN
	case "dana":
		return PAYMENT_DANA
	case "ovo":
		return PAYMENT_OVO
	case "gopay":
		return PAYMENT_GOPAY
	case "shopeepay":
		return PAYMENT_SHOPEEPAY
	case "qris", "qris_midtrans", "qris_harsya":
		return PAYMENT_QRIS
	case "va_bca":
		return PAYMENT_VA_BCA
	case "va_bni":
		return PAYMENT_VA_BNI
	case "va_bri":
		return PAYMENT_VA_BRI
	case "va_mandiri":
		return PAYMENT_VA_MANDIRI
	case "va_permata":
		return PAYMENT_VA_PERMATA
	case "va_sinarmas":
		return PAYMENT_VA_SINARMAS
	}

	// Then check endpoint patterns
	for pattern, method := range endpointMap {
		if strings.Contains(endpoint, pattern) {
			return method
		}
	}

	return "unknown"
}
