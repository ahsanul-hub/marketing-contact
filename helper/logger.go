package helper

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// LogLevelString maps log levels to their string representations
var LogLevelString = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// LogLevelColor maps log levels to ANSI color codes
var LogLevelColor = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

const resetColor = "\033[0m"

// Logger provides structured logging functionality
type Logger struct {
	prefix string
}

// NewLogger creates a new logger instance with optional prefix
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// formatMessage formats the log message with timestamp, level, and prefix
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := LogLevelString[level]
	color := LogLevelColor[level]

	if l.prefix != "" {
		return fmt.Sprintf("%s[%s] %s%s %s[%s]%s %s",
			color, timestamp, levelStr, resetColor,
			color, l.prefix, resetColor, message)
	}

	return fmt.Sprintf("%s[%s] %s%s %s",
		color, timestamp, levelStr, resetColor, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println(l.formatMessage(DEBUG, message))
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println(l.formatMessage(INFO, message))
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println(l.formatMessage(WARN, message))
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println(l.formatMessage(ERROR, message))
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println(l.formatMessage(FATAL, message))
	os.Exit(1)
}

// Section creates a section header in logs
func (l *Logger) Section(title string) {
	log.Println("")
	log.Printf("=== %s ===", title)
}

// SubSection creates a subsection header in logs
func (l *Logger) SubSection(title string) {
	log.Printf("--- %s ---", title)
}

// EndSection marks the end of a section
func (l *Logger) EndSection() {
	log.Println("=== END SECTION ===")
	log.Println("")
}

// Data logs structured data in a readable format
func (l *Logger) Data(title string, data interface{}) {
	l.Section(title)

	switch v := data.(type) {
	case []interface{}:
		for i, item := range v {
			log.Printf("[%d] %+v", i+1, item)
		}
	case map[string]interface{}:
		for key, value := range v {
			log.Printf("%s: %+v", key, value)
		}
	default:
		log.Printf("%+v", data)
	}

	l.EndSection()
}

// Table logs data in a table-like format
func (l *Logger) Table(headers []string, rows [][]string) {
	l.Section("TABLE DATA")

	// Print headers
	headerStr := "| "
	for _, header := range headers {
		headerStr += fmt.Sprintf("%-15s | ", header)
	}
	log.Println(headerStr)

	// Print separator
	separator := "| "
	for range headers {
		separator += "---------------- | "
	}
	log.Println(separator)

	// Print rows
	for i, row := range rows {
		rowStr := "| "
		for _, cell := range row {
			rowStr += fmt.Sprintf("%-15s | ", cell)
		}
		log.Printf("[%d] %s", i+1, rowStr)
	}

	l.EndSection()
}

// PaymentMethodData logs payment method information in a structured format
func (l *Logger) PaymentMethodData(paymentMethods []interface{}) {
	l.Section("PAYMENT METHODS")

	for i, pm := range paymentMethods {
		log.Printf("[%d] Payment Method: %+v", i+1, pm)
	}

	l.EndSection()
}

// SettlementData logs settlement information in a structured format
func (l *Logger) SettlementData(settlements []interface{}) {
	l.Section("SETTLEMENTS")

	for i, settlement := range settlements {
		log.Printf("[%d] Settlement: %+v", i+1, settlement)
	}

	l.EndSection()
}

// RouteWeightData logs route weight information in a structured format
func (l *Logger) RouteWeightData(weights []interface{}) {
	l.Section("CHANNEL ROUTE WEIGHTS")

	for i, weight := range weights {
		log.Printf("[%d] Route Weight: %+v", i+1, weight)
	}

	l.EndSection()
}

// Global logger instance
var AppLogger = NewLogger("DCB-BE")

// Convenience functions for global logging
func Debug(format string, args ...interface{}) {
	AppLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	AppLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	AppLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	AppLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	AppLogger.Fatal(format, args...)
}

func Section(title string) {
	AppLogger.Section(title)
}

func SubSection(title string) {
	AppLogger.SubSection(title)
}

func EndSection() {
	AppLogger.EndSection()
}

func Data(title string, data interface{}) {
	AppLogger.Data(title, data)
}

func Table(headers []string, rows [][]string) {
	AppLogger.Table(headers, rows)
}
